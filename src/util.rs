use failure::{format_err, Error};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;

#[derive(Debug, Deserialize, Serialize)]
pub struct Info {
    pub contest_url: Option<String>,
    pub problem_url: Option<String>,
}

pub fn algpath() -> PathBuf {
    dirs::home_dir().unwrap().join("Programs").join("Algorithm")
}

pub fn online(path: &Path) -> Result<bool, Error> {
    if !path.is_dir() {
        return Ok(false);
    }

    let info: Info = toml::from_str(&fs::read_to_string(path.join("info.toml"))?)?;
    Ok(info.contest_url.is_some())
}

pub fn source(path: &Path) -> Option<PathBuf> {
    let pred = if path.is_dir() {
        PathBuf::from(path).join("main.cpp")
    } else {
        PathBuf::from(path)
    };
    if pred.exists() {
        Some(pred)
    } else {
        None
    }
}

pub fn combine(src: &Path) -> Result<PathBuf, Error> {
    let mut out_src = PathBuf::from(src);
    out_src.set_file_name(format!(
        "{}_combined.cpp",
        out_src.file_stem().unwrap().to_str().unwrap()
    ));
    let process = Command::new(
        algpath()
            .join("expander")
            .join("expander.py")
            .to_str()
            .unwrap(),
    )
    .arg(src)
    .arg(&out_src)
    .spawn()?
    .wait()?;
    if !process.success() {
        Err(format_err!("Failed to combine"))
    } else {
        Ok(out_src)
    }
}
