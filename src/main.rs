extern crate clap;
extern crate dirs;
extern crate glob;
#[macro_use]
extern crate serde_derive;
extern crate toml;

use clap::{App, Arg, SubCommand};
use env_logger;
use log::{info, warn};
use std::env;
use std::ffi::OsString;
use std::fs;
use std::fs::File;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::time::Instant;

#[derive(Debug, Deserialize, Serialize)]
struct Info {
    contest_url: Option<String>,
}

fn command_init(matches: &clap::ArgMatches) {
    let url = matches.value_of("url").unwrap();
    let problems = matches.values_of("problems").unwrap();
    let code = "
import sys
import onlinejudge

contest = onlinejudge.dispatch.contest_from_url(sys.argv[1])
if isinstance(contest, onlinejudge.service.atcoder.AtCoderContest):
    print('atcoder-' + contest.contest_id)
    exit(0)
exit(1)
    ";

    let result = Command::new("python3")
        .args(&["-c", code])
        .arg(url)
        .output()
        .expect("Fail to expand");
    let dir = String::from_utf8(result.stdout).expect("Fail to get dir");
    let dir = dir.trim();
    assert!(dir != "");
    info!("base dir: {}", dir);
    let dir = PathBuf::from(dir);
    fs::create_dir(&dir).expect("Fail make dir");
    for prob in problems {
        let pdir = dir.clone().join(prob);
        info!("make: {:?}", pdir);
        fs::create_dir(&pdir).expect("Create problem dir");
        fs::copy(
            algpath().join("src").join("base.cpp"),
            pdir.join("main.cpp"),
        )
        .expect("Copy main.cpp");
        fs::create_dir(pdir.join("ourtest")).expect("Create test dir");
    }
    let info = Info {
        contest_url: Some(url.to_string()),
    };
    let toml = toml::to_string(&info).unwrap();
    fs::write(dir.join("info.toml"), toml).expect("write toml");
}

fn algpath() -> PathBuf {
    dirs::home_dir().unwrap().join("Programs").join("Algorithm")
}

fn build(src: &Path) {
    info!("build: {:?}", src);
    let mut process = Command::new(env::var_os("CXX").unwrap_or(OsString::from("g++")))
        .arg("-std=c++17")
        // warnings
        .args(&[
            "-Wall",
            "-Wextra",
            "-Wshadow",
            "-Wconversion",
            "-Wno-sign-conversion",
        ])
        // debug, sanitize
        .arg("-g")
        .arg("-fsanitize=address,undefined")
        .arg("-fno-omit-frame-pointer")
        .arg("-DLOCAL")
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
        .spawn()
        .expect("Fail to compile");
    let status = process.wait().expect("Failed to compile");
    assert!(status.success());
}

fn command_build(matches: &clap::ArgMatches) {
    let src = matches.value_of("source").unwrap();
    let path = PathBuf::from(src);
    let path = if path.is_dir() {
        path.join("main.cpp")
    } else {
        path
    };
    info!("{:?}", path);
    build(path.as_path());
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
    return true;
}
fn get_url(url: &str, problem: &str) -> Option<String> {
    let code = "
import sys
import onlinejudge
url = sys.argv[1]    
problem = sys.argv[2]
contest = onlinejudge.dispatch.contest_from_url(url)
if not contest:
    sys.exit(1)
list = contest.list_problems()
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

fn test(dir: &str, test: &str) {
    let test_dir = PathBuf::from(dir).join(test);
    if !test_dir.exists() {
        info!("No test dir: {:?}", test_dir);
        return;
    }
    let mut files = Vec::<PathBuf>::new();
    for entry in fs::read_dir(PathBuf::from(dir).join(test)).unwrap() {
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
        let bin = fs::canonicalize(PathBuf::from(dir).join("main")).expect("Fail");
        let mut output = f.clone();
        output.set_file_name(format!("{}.out", f.file_stem().unwrap().to_str().unwrap()));
        let start = Instant::now();
        let command = Command::new(bin)
            .stdin(File::open(f).expect("Fail input"))
            .stderr(Stdio::inherit())
            .output()
            .expect("Fail to compile");
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
                } else {
                    info!("AC");
                }
            }
            Err(_) => {
                info!("No answer file");
                println!("=== output: ===");
                print!("{}", actual);
            }
        }
        info!("Time: {} ms", end.as_millis())
    }
}
fn command_test(matches: &clap::ArgMatches) {
    let dir = matches.value_of("problem").unwrap();
    build(PathBuf::from(dir).join("main.cpp").as_path());

    let info: Info =
        toml::from_str(&fs::read_to_string("info.toml").expect("Fail to read info.toml"))
            .expect("Fail to read toml");
    if let Some(url) = info.contest_url {
        let test_dir = PathBuf::from(dir).join("test");
        if !test_dir.exists() {
            info!("download: {}", url);
            let status = Command::new("oj")
                .current_dir(dir)
                .arg("d")
                .arg(get_url(&url, dir).expect("fail to get url"))
                .spawn()
                .expect("Fail to expand")
                .wait()
                .expect("Fail to expand");
            if !status.success() {
                warn!("Failed to download case");
            }
        }
        test(dir, "test")
    }
    test(dir, "ourtest");
}

fn combine(src: &Path) {
    let mut out_src = PathBuf::from(src);
    out_src.set_file_name(format!(
        "{}_combined.cpp",
        out_src.file_stem().expect("fail").to_str().unwrap()
    ));
    let status = Command::new(
        algpath()
            .join("expander")
            .join("expander.py")
            .to_str()
            .expect("fail"),
    )
    .arg(src)
    .arg(out_src)
    .spawn()
    .expect("Fail to expand")
    .wait()
    .expect("Fail to expand");
    assert!(status.success());
}

fn command_submit(matches: &clap::ArgMatches) {
    let dir = matches.value_of("problem").unwrap();
    if !Path::new(dir).is_dir() {
        combine(&Path::new(dir));
        return;
    }
    let info: Info =
        toml::from_str(&fs::read_to_string("info.toml").expect("Fail to read info.toml"))
            .expect("Fail to read toml");
    combine(&PathBuf::from(dir).join("main.cpp"));
    if let Some(url) = info.contest_url {
        info!("submit: {}", url);
        Command::new("oj")
            .current_dir(dir)
            .arg("s")
            .arg("--no-open")
            .args(&["-w", "0"])
            .arg("main_combined.cpp")
            .spawn()
            .expect("Fail to expand")
            .wait()
            .expect("Fail to expand");
    }
}

fn main() {
    env::set_var("RUST_LOG", "info");
    env_logger::Builder::from_default_env()
        .format_timestamp(None)
        .format_module_path(false)
        .init();
    let app = App::new("clapex")
        .subcommand(
            SubCommand::with_name("i")
                .about("init")
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
                .arg(Arg::with_name("source").help("source").required(true)),
        )
        .subcommand(
            SubCommand::with_name("t")
                .about("test")
                .arg(Arg::with_name("problem").help("problem").required(true)),
        )
        .subcommand(
            SubCommand::with_name("s")
                .about("submit")
                .arg(Arg::with_name("problem").help("problem").required(true)),
        );
    let matches = app.get_matches();
    if let Some(ref matches) = matches.subcommand_matches("i") {
        command_init(matches)
    }
    if let Some(ref matches) = matches.subcommand_matches("b") {
        command_build(matches)
    }
    if let Some(ref matches) = matches.subcommand_matches("t") {
        command_test(matches)
    }
    if let Some(ref matches) = matches.subcommand_matches("s") {
        command_submit(matches)
    }
}
