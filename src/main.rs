mod util;

#[macro_use]
extern crate serde_derive;

use url::{Url};
use clap::{App, Arg, SubCommand};
use env_logger;
use failure::{format_err, Error};
use log::{info, warn};
use std::env;
use std::ffi::OsString;
use std::fs;
use std::fs::File;
use std::os::unix::process::CommandExt;
use std::io::{stdout, Write};
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::time::Instant;
use util::*;

fn problem_init(pdir: &Path, contest_url: &Option<String>) -> Result<(), Error> {
    info!("init problem: {:?}", pdir);
    info!("make problem dir");
    fs::create_dir(&pdir)?;
    fs::copy(
        algpath().join("src").join("base.cpp"), // TODO: default source path
        pdir.join("main.cpp"),
    )?;
    fs::create_dir(pdir.join("ourtest"))?;
    let info = Info {
        contest_url: contest_url.clone(),
        problem_url: None,
    };
    let toml = toml::to_string(&info)?;
    fs::write(pdir.join("info.toml"), toml)?;
    Ok(())
}

fn command_init(matches: &clap::ArgMatches) -> Result<(), Error> {
    let url = matches.value_of("url").unwrap();
    let problems = matches.values_of("problems").unwrap();
    info!("contest init: {:?} {:?}", url, problems);
    let code = "
import sys
import onlinejudge

contest = onlinejudge.dispatch.contest_from_url(sys.argv[1])
if isinstance(contest, onlinejudge.service.atcoder.AtCoderContest):
    print('atcoder-{}'.format(contest.contest_id))
    exit(0)
if isinstance(contest, onlinejudge.service.codeforces.CodeforcesContest):
    print('codeforces-{}'.format(contest.contest_id))
    exit(0)
exit(1)
    ";
    let (dir, url) = if Url::parse(url).is_ok() {
        let result = Command::new("python3")
            .args(&["-c", code])
            .arg(url)
            .output()?;
        (String::from_utf8(result.stdout)?, Some(String::from(url)))
    } else {
        info!("Offline contest");
        (String::from(url), None)
    };
    let dir = dir.trim();
    assert!(dir != "");
    let dir = PathBuf::from(dir);
    info!("contest dir: {:?}", dir);
    fs::create_dir(&dir)?;
    for prob in problems {
        let pdir = dir.clone().join(prob);
        problem_init(&pdir, &url)?;
    }
    Ok(())
}

fn build(src: &Path, opt: bool) -> Result<(), Error> {
    info!("build: {:?}", src);
    let cxx = env::var_os("CXX").unwrap_or(OsString::from("g++"));
    let mut process = Command::new(cxx);
    process
        .arg("-std=c++17")
        // warnings
        .args(&[
            "-Wall",
            "-Wextra",
            "-Wshadow",
            "-Wconversion",
            "-Wno-sign-conversion",
        ])
        .arg("-g")
        .arg("-DLOCAL");
    
    if !opt {
        // debug, sanitize
        process
            .arg("-fsanitize=address,undefined")
            .arg("-fno-omit-frame-pointer")
    } else {
        info!("opt build");
        process.arg("-O2")
    };
    let process = process
        // include
        .args(&["-I", algpath().join("src").to_str().unwrap()])
        // output, source
        .args(&[
            "-o",
            src.parent()
                .unwrap()
                .join(src.file_stem().unwrap())
                .to_str()
                .unwrap(),
        ])
        .arg(src.as_os_str().to_str().unwrap())
        .spawn()?
        .wait()?;
    if !process.success() {
        Err(format_err!("Failed to compile"))
    } else {
        Ok(())
    }
}

fn command_build(matches: &clap::ArgMatches) -> Result<(), Error> {
    let src = matches.value_of("source").unwrap();
    let opt = matches.is_present("opt");
    match source(Path::new(&src)) {
        Some(src) => build(&src, opt),
        None => Err(format_err!("No source file")),
    }
}

fn command_run(matches: &clap::ArgMatches) -> Result<(), Error> {
    let src = matches.value_of("source").unwrap();
    let opt = matches.is_present("opt");
    match source(Path::new(&src)) {
        Some(src) => build(&src, opt)?,
        None => return Err(format_err!("No source file")),
    }
    let bin = fs::canonicalize(PathBuf::from(src).join("main"))?;
    Command::new(bin).exec();
    Err(format_err!("Failed to exec"))
}

fn check_diff(actual: &str, expect: &str) -> bool {
    let mut actual_lines = actual.lines();
    for expect_line in expect.lines() {
        if let Some(actual_line) = actual_lines.next() {
            if actual_line.trim_end() == expect_line.trim_end() {
                continue;
            }
        }
        if expect_line.len() == 0 {
            continue;
        }
        return false;
    }
    for actual_line in actual_lines {
        if actual_line.trim_end().len() > 0 {
            return false;
        }
    }
    return true;
}
fn get_url(url: &str, problem: &str) -> Option<String> {
    let code = "
import sys
import onlinejudge
import onlinejudge._implementation.utils as utils

url = sys.argv[1]
problem = sys.argv[2]
contest = onlinejudge.dispatch.contest_from_url(url)
if not contest:
    sys.exit(1)
with utils.with_cookiejar(utils.new_session_with_our_user_agent(), path=utils.default_cookie_path) as sess:
    list = contest.list_problems(session=sess)
for p in list:
    purl = p.get_url()
    if purl.lower().endswith(problem.lower()):
        print(purl)
        sys.exit(0)
if len(problem) == 1:
    id = ord(problem.lower()) - ord('a')
    if id < len(list):
        print(list[id].get_url())
        sys.exit(0)
sys.exit(1)    
    ";

    let result = Command::new("python3")
        .args(&["-c", code])
        .arg(url)
        .arg(problem)
        .output()
        .expect("Fail to expand");
    if !result.status.success() {
        warn!("Failed to get url: {} {}", url, problem);
        return None;
    }
    match String::from_utf8(result.stdout) {
        Ok(v) => Some(v),
        Err(_) => None,
    }
}

fn test(dir: &Path, test_dir: &Path) -> Result<(), Error> {
    if !test_dir.exists() {
        return Err(format_err!("No test dir: {:?}", test_dir));
    }
    let mut files = Vec::<PathBuf>::new();
    for entry in fs::read_dir(test_dir).unwrap() {
        let entry = entry.unwrap();
        let path: PathBuf = entry.path();
        if let Some(ext) = path.extension() {
            if ext == "in" {
                files.push(path)
            }
        }
    }
    files.sort();
    for f in files {
        info!("test: {:?}", f);
        let bin = fs::canonicalize(PathBuf::from(dir).join("main"))?;
        let mut output = f.clone();
        output.set_file_name(format!("{}.out", f.file_stem().unwrap().to_str().unwrap()));
        let start = Instant::now();
        let command = Command::new(bin)
            .stdin(File::open(f)?)
            .stderr(Stdio::inherit())
            .output()?;
        let end = start.elapsed();
        let actual = String::from_utf8(command.stdout).unwrap();
        match fs::read_to_string(output) {
            Ok(expect) => {
                if !check_diff(&actual, &expect) {
                    warn!("WA");
                    println!("=== output: ===");
                    print!("{}", actual);
                    println!("=== expect: ===");
                    print!("{}", expect);
                    stdout().flush().unwrap();
                } else {
                    info!("AC");
                }
            }
            Err(_) => {
                info!("No answer file");
                println!("=== output: ===");
                print!("{}", actual);
                stdout().flush()?;
            }
        }
        info!("Time: {} ms", end.as_millis())
    }
    Ok(())
}
fn command_test(matches: &clap::ArgMatches) -> Result<(), Error> {
    let pdir = matches.value_of("problem").unwrap();
    let opt = matches.is_present("opt");
    let pdir = PathBuf::from(pdir);
    let pname = pdir.file_name().unwrap().to_str().unwrap();

    build(pdir.join("main.cpp").as_path(), opt)?;

    let info: Info = toml::from_str(&fs::read_to_string(pdir.join("info.toml"))?)?;
    if let Some(url) = info.contest_url {
        let test_dir = pdir.join("test");
        if !test_dir.exists() {
            info!("download: {}", url);
            let status = Command::new("oj")
                .current_dir(&pdir)
                .arg("d")
                .arg(get_url(&url, pname).expect("Get problem url"))
                .spawn()?
                .wait()?;
            if !status.success() {
                warn!("Failed to download case");
            }
        }
        test(&pdir, &pdir.join("test"))?;
    }
    test(&pdir, &pdir.join("ourtest"))
}

fn command_submit(matches: &clap::ArgMatches) -> Result<(), Error> {
    let pdir = matches.value_of("problem").unwrap();
    let pdir = PathBuf::from(pdir);
    let out_src = match source(Path::new(&pdir)) {
        Some(src) => combine(&src)?,
        None => return Err(format_err!("No source file")),
    };
    if matches.is_present("clip") {
        info!("copy to clipboard");
        Command::new("pbcopy")
            .stdin(File::open(&out_src)?)
            .spawn()?
            .wait()?;
    }
    if !online(&pdir)? {
        return Ok(());
    }
    let info: Info = toml::from_str(&fs::read_to_string(pdir.join("info.toml"))?)?;
    combine(&pdir.join("main.cpp"))?;
    if let Some(url) = info.contest_url {
        info!("submit: {}", url);
        Command::new("oj")
            .current_dir(&pdir)
            .arg("s")
            .arg("--no-open")
            .args(&["-w", "0"])
            .arg("main_combined.cpp")
            .spawn()?
            .wait()?;
    }
    Ok(())
}

fn main() {
    env::set_var("RUST_LOG", "info");
    env_logger::Builder::from_default_env()
        .format_timestamp(None)
        .format_module_path(false)
        .init();
    let app = App::new("supporter")
        .subcommand(
            SubCommand::with_name("i")
                .about("contest init")
                .arg(Arg::with_name("url").help("url").required(true))
                .arg(
                    Arg::with_name("problems")
                        .multiple(true)
                        .takes_value(true)
                        .required(true),
                ),
        )
        .subcommand(
            SubCommand::with_name("b")
                .about("build")
                .arg(Arg::with_name("source").help("source").required(true))
                .arg(Arg::with_name("opt").help("optimize").short("O")),
        )
        .subcommand(
            SubCommand::with_name("r")
                .about("run")
                .arg(Arg::with_name("source").help("source").required(true))
                .arg(Arg::with_name("opt").help("optimize").short("O")),
        )
        .subcommand(
            SubCommand::with_name("t")
                .about("test")
                .arg(Arg::with_name("problem").help("problem").required(true))
                .arg(Arg::with_name("opt").help("optimize").short("O")),
        )
        .subcommand(
            SubCommand::with_name("s")
                .about("submit")
                .arg(Arg::with_name("problem").help("problem").required(true))
                .arg(Arg::with_name("clip").help("copy to clipboard").short("c")),
        );
    let matches = app.get_matches();
    if let Some(ref matches) = matches.subcommand_matches("i") {
        command_init(matches).expect("Contest init")
    }
    if let Some(ref matches) = matches.subcommand_matches("b") {
        command_build(matches).expect("Build")
    }
    if let Some(ref matches) = matches.subcommand_matches("r") {
        command_run(matches).expect("Run")
    }
    if let Some(ref matches) = matches.subcommand_matches("t") {
        command_test(matches).expect("Test")
    }
    if let Some(ref matches) = matches.subcommand_matches("s") {
        command_submit(matches).expect("Submit")
    }
}
