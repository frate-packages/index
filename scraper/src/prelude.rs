use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Debug, Eq, PartialEq, Hash)]
#[allow(non_snake_case)]
pub struct Package {
    pub name: String,
    pub git: Option<String>,
}
