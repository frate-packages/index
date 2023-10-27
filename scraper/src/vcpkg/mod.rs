use crate::prelude::Package;
use clap::{Args, Parser};
use reqwest::blocking::get;
use serde::{Deserialize, Serialize};
use std::io::Write;

#[derive(Serialize, Deserialize, Debug)]
#[allow(non_snake_case)]
pub struct Item {
    pub Name: String,
    pub Homepage: Option<String>,
}

#[derive(Serialize, Deserialize, Debug)]
#[allow(non_snake_case)]
pub struct Vcpkg {
    Baseline: Option<String>,
    Size: Option<usize>,
    Source: Vec<Item>,
}

impl Vcpkg {
    pub fn new() -> Self {
        Self {
            Baseline: None,
            Size: None,
            Source: Vec::new(),
        }
    }

    pub fn scrape(
        &mut self,
        //filename: Option<String>,
    ) -> Result<Vec<Package>, Box<dyn std::error::Error>> {
        //
        // default sample.json filename
        //let mut output_file_name = String::from("../index.json");

        //// use defined ouput filename
        //if let Some(out_name) = filename {
        //    output_file_name = out_name
        //}

        // Request json file
        println!("[#] Downloading output.json");

        // serialize data into struct Vcpkg
        let resp: Vcpkg = get("https://vcpkg.io/output.json")?.json()?;
        println!("[#] Download complete");

        // create vec for filtering
        let mut filtered = Vec::new();

        // filter responce if Homepage exists and is a github link
        println!("[#] Sorting vcpkg...");
        for item in resp.Source.into_iter() {
            println!("[#] Checking: {}", item.Name);
            // if Homepage exists
            if let Some(val) = item.Homepage.clone() {
                // if link contains the string github
                if val.contains("github") {
                    // pack it into a new package
                    let package = Package {
                        name: item.Name,
                        git: item.Homepage,
                    };
                    // push package to filtered stack
                    filtered.push(package);
                }
            }
        }

        //jump back a directory and create a folder with filtered Name and save a json with Name,
        //and git link in the directory
        //  for item in filtered.iter() {
        //      let mut path = String::from("../");
        //      path.push_str(&item.Name);
        //      std::fs::create_dir_all(path.clone())?;
        //      path.push_str("/info.json");
        //      let mut file = std::fs::File::create(path)?;
        //      let string = serde_json::to_string(item)?;
        //      file.write_all(string.as_bytes())?;
        //  }

        //  // stringify json and write to file
        //  println!("[#] Writing to file: {}", output_file_name);
        //  if let Ok(string) = serde_json::to_string(&filtered) {
        //      match std::fs::write(output_file_name, string) {
        //          Ok(_) => {
        //              println!("[#] File Succesfully written");
        //          }
        //          Err(e) => {
        //              println!("[!] Could not write file for some reason.\n {}", e);
        //          }
        //      }
        //  }

        Ok(filtered)
    }
}
