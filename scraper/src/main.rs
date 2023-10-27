#![allow(unused)]
mod conan;
mod prelude;
mod vcpkg;

use std::{collections::HashSet, io::Write};

use conan::Conan;
use vcpkg::Vcpkg;

use clap::Parser;
use serde::{Deserialize, Serialize};

#[derive(Parser, Debug)]
#[command(author = "Brainfart", version, about, long_about = None)]
struct Args {
    /// Output directory
    #[arg(short, action)]
    output_directory: String,

    /// Scrape Vcpkg
    #[arg(short, action)]
    vcpkg: bool,

    /// Scrape Conan
    #[arg(short, action)]
    conan: bool,
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = Args::parse();

    let mut set = HashSet::new();

    let mut vcpkg = Vcpkg::new();
    let mut conan = Conan::new();

    let mut vcpkg_data = Vec::new();
    let mut conan_data = Vec::new();

    if args.vcpkg {
        vcpkg_data = vcpkg.scrape()?;
        for data in vcpkg_data {
            set.insert(data);
        }
    }

    if args.conan {
        conan_data = conan.scrape()?;
        for data in conan_data {
            set.insert(data);
        }
    }

    // iter over set if dir with name does not exist,
    // create it and add info.json with package data

    println!("[#] Creating directories");

    for data in set.iter() {
        if let Ok(_) = std::fs::create_dir_all(format!("{}/{}", args.output_directory, data.name)) {
            println!("[#] Creating: {}", data.name);
            let mut file = std::fs::File::create(format!(
                "{}/{}/info.json",
                args.output_directory, data.name
            ))?;
            let string = serde_json::to_string(data)?;
            file.write_all(string.as_bytes());
        }
    }
    Ok(())
}
