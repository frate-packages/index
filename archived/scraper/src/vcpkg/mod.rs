use crate::prelude::Package;
use clap::{Args, Parser};
use core::fmt;
use reqwest::blocking::get;
use serde::{
    de::{Error, Visitor},
    Deserialize, Deserializer, Serialize,
};
use std::io::Write;

#[allow(non_snake_case)]
#[derive(Serialize, Deserialize, Debug)]
#[serde(untagged)]
pub enum Descriptions {
    Arr(Vec<String>),
    Str(String),
}

#[derive(Serialize, Deserialize, Debug)]
#[allow(non_snake_case)]
pub struct Item {
    pub Name: String,
    pub Homepage: Option<String>,
    //#[serde(skip)]
    //pub Description: Descriptions,
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
                        description: String::new(), //description: match item.Description {
                                                    //    Descriptions::Arr(v) => v,
                                                    //    Descriptions::Str(v) => vec![v],
                                                    //},
                    };
                    // push package to filtered stack
                    filtered.push(package);
                }
            }
        }

        Ok(filtered)
    }
}

//pub fn deserialize<'de, D>(deserializer: D) -> Result<Descriptions, D::Error>
//where
//    D: Deserializer<'de>,
//{
//    struct KeyVisitor;
//
//    impl<'de> Visitor<'de> for KeyVisitor {
//        type Value = Descriptions;
//
//        fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
//            formatter.write_str("a single string or an array of strings")
//        }
//
//        fn visit_str<E>(self, value: &str) -> Result<Self::Value, E>
//        where
//            E: Error,
//        {
//            Ok(Descriptions::Str(value.to_owned()))
//        }
//
//        fn visit_seq<A>(self, mut seq: A) -> Result<Self::Value, A::Error>
//        where
//            A: serde::de::SeqAccess<'de>,
//        {
//            let mut values = Vec::new();
//            while let Some(value) = seq.next_element()? {
//                values.push(value);
//            }
//            Ok(Descriptions::Arr(values))
//        }
//    }
//
//    deserializer.deserialize_any(KeyVisitor)
//}
