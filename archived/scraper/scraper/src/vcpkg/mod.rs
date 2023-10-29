mod serializer;
use crate::prelude::Package;
use clap::{Args, Parser};
use core::fmt;
use reqwest::blocking::get;
use serde::{
    de::{Error, Visitor},
    Deserialize, Deserializer, Serialize, Serializer,
};
use std::io::Write;

#[allow(non_snake_case)]
#[derive(Serialize, Deserialize, Debug)]
#[serde(untagged)]
pub enum Descriptions {
    Vec(Vec<String>),
    String(String),
}

#[derive(Serialize, Deserialize, Debug)]
#[allow(non_snake_case)]
pub struct Vcpkg {
    Baseline: Option<String>,
    Size: Option<usize>,
    pub Source: Vec<Item>,
}

#[derive(Serialize, Deserialize, Debug)]
#[allow(non_snake_case)]
pub struct Item {
    pub Name: String,
    pub Homepage: Option<String>,
    pub Description: Option<Descriptions>,
    //#[serde(skip)]
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
                        description: match item.Description.unwrap() {
                            Descriptions::Vec(v) => v.join(" "),
                            Descriptions::String(s) => s,
                        },
                    };
                    // push package to filtered stack
                    filtered.push(package);
                }
            }
        }

        Ok(filtered)
    }
}
