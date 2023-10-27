#![allow(unused)]
use std::{
    collections::HashMap,
    io::Result,
    sync::{Arc, Mutex},
    thread::{self, JoinHandle},
    time::Duration,
};

use crate::prelude::Package;

use reqwest::blocking::get;
use scraper::*;

use serde::{Deserialize, Serialize};
use serde_json::*;

const URL: &'static str = "https://conan.io/center/recipes";

#[derive(Debug, Serialize, Deserialize, Clone)]
struct Item {
    name: String,
    href: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Conan {}
impl Conan {
    pub fn new() -> Self {
        Self {}
    }
    pub fn scrape(&self) -> core::result::Result<Vec<Package>, Box<dyn std::error::Error>> {
        // create conan struct
        let mut data = Vec::new();

        // setup a thread safe array to store scraped data
        let mut t_array: Arc<Mutex<Vec<Item>>> = Arc::new(Mutex::new(Vec::new()));

        // scrape the main page to get all page names;
        let html = get(URL).unwrap().text().unwrap();
        let doc = Html::parse_document(html.as_str());
        let list_selector = Selector::parse(".list-group .list-group-item").unwrap();

        // Create a vector to store the scaped elements so that it can be scraped in different threads.
        let mut elements = Vec::new();
        for element in doc.select(&list_selector) {
            elements.push(element.html().clone());
        }

        // create a vector that can contain all the threads, so it can be joined later on;
        let mut threads = Vec::new();

        println!("[#] Scraping: Conan...");
        // loop through the elements and spawn a thread to scrape the specific package
        for element in elements {
            // create a clone of the ARC pointer and pass it into the thread;
            let t_array_clone = t_array.clone();

            // sleep for 50ms before spanning a new thread
            thread::sleep(Duration::from_millis(100));

            threads.push(thread::spawn(move || {
                // selector link of package
                let link_selector = Selector::parse("a").unwrap();

                // selector name of package
                let name_selector = Selector::parse("a h3").unwrap();

                // parse the package page html
                let doc = Html::parse_document(&element.clone());

                // select the github href
                let link = doc
                    .select(&link_selector.clone())
                    .next()
                    .unwrap()
                    .value()
                    .attr("href");

                // select the name, and remove exccess crap;
                let name = doc
                    .select(&name_selector.clone())
                    .next()
                    .unwrap()
                    .inner_html()
                    .replace("<!-- -->/<!-- -->", " ")
                    .split(" ")
                    .next()
                    .unwrap()
                    .to_owned();

                // pack data into Package struct;
                let mut item = Item {
                    name,
                    href: link.unwrap().to_owned(),
                };

                let mut retry_couter: HashMap<String, usize> = HashMap::new();

                scrape(&mut item, &mut retry_couter);
                {
                    let mut lock = t_array_clone.lock().unwrap();
                    let item = Item {
                        name: item.name,
                        href: item.href,
                    };
                    lock.push(item);
                }
            }));
        }
        // join all threads;
        for t in threads {
            t.join();
        }

        // transfere data from thread safe array to normal data structure
        let lock = t_array.lock().unwrap().clone();

        // repack mutext into package;
        data = lock
            .into_iter()
            .map(|it| Package {
                name: it.name,
                git: Some(it.href),
            })
            .collect::<Vec<Package>>();

        Ok(data)
    }
}
fn scrape(item: &mut Item, retry_couter: &mut HashMap<String, usize>) {
    //let mut Item = Item.clone();

    // get html from package page;
    if let Ok(html) = get(format!("https://conan.io{}", item.href)) {
        let html = html.text().unwrap();
        let github_link_selector = Selector::parse("a").unwrap();
        let doc = Html::parse_document(html.as_str());

        // get the href of the github page;
        // filter out only github and non conan links
        for link in doc.select(&github_link_selector) {
            if let Some(l) = link.value().attr("href") {
                if l.contains("github") && !l.contains("conan") {
                    item.href = l.to_owned();
                    println!("[#] Checking: {}", item.name);
                } else {
                }
            } else {
            }
        }
    } else {
        println!("[!] Could not load page: {}, retry in 500ms", item.name);
        if let Some(count) = retry_couter.get_mut(&item.name) {
            if *count < 5 {
                *count += 1;
            } else {
                return;
            }
        } else {
            retry_couter.insert(item.name.to_string(), 1);
        }
        std::thread::sleep(Duration::from_millis(500));
        scrape(item, retry_couter);
    }
}
