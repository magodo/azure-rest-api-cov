resource "azurerm_resource_group" "root" {
  location = "eastus2"
  name     = "azure-rest-api-cov"
}

resource "azurerm_cosmosdb_account" "root" {
  location            = "eastus2"
  name                = "azure-rest-api-cov"
  offer_type          = "Standard"
  resource_group_name = azurerm_resource_group.root.name
  capabilities {
    name = "EnableServerless"
  }
  consistency_policy {
    consistency_level = "Strong"
  }
  geo_location {
    failover_priority = 0
    location          = "eastus2"
    zone_redundant    = true
  }
  backup {
    type = "Continuous"
  }
}

resource "azurerm_cosmosdb_sql_database" "terraform" {
  name                = "terraform"
  resource_group_name = azurerm_cosmosdb_account.root.resource_group_name
  account_name        = azurerm_cosmosdb_account.root.name
}

resource "azurerm_cosmosdb_sql_container" "build" {
  name                  = "build"
  resource_group_name   = azurerm_cosmosdb_sql_database.terraform.resource_group_name
  account_name          = azurerm_cosmosdb_sql_database.terraform.account_name
  database_name         = azurerm_cosmosdb_sql_database.terraform.name
  partition_key_path    = "/id"
  partition_key_version = 2
}

resource "azurerm_cosmosdb_sql_container" "rp" {
  name                  = "rp"
  resource_group_name   = azurerm_cosmosdb_sql_database.terraform.resource_group_name
  account_name          = azurerm_cosmosdb_sql_database.terraform.account_name
  database_name         = azurerm_cosmosdb_sql_database.terraform.name
  partition_key_path    = "/buildId"
  partition_key_version = 2
  unique_key {
    paths = ["/name"]
  }
}

resource "azurerm_cosmosdb_sql_container" "path" {
  name                  = "path"
  resource_group_name   = azurerm_cosmosdb_sql_database.terraform.resource_group_name
  account_name          = azurerm_cosmosdb_sql_database.terraform.account_name
  database_name         = azurerm_cosmosdb_sql_database.terraform.name
  partition_key_path    = "/buildId"
  partition_key_version = 2
  unique_key {
    paths = ["/rpName", "/name"]
  }
}

resource "azurerm_cosmosdb_sql_container" "path_grant" {
  name                  = "path_grant"
  resource_group_name   = azurerm_cosmosdb_sql_database.terraform.resource_group_name
  account_name          = azurerm_cosmosdb_sql_database.terraform.account_name
  database_name         = azurerm_cosmosdb_sql_database.terraform.name
  partition_key_path    = "/resourceProvider"
  partition_key_version = 2
  unique_key {
    paths = ["/path"]
  }
}

resource "azurerm_cosmosdb_sql_container" "prop_grant" {
  name                  = "prop_grant"
  resource_group_name   = azurerm_cosmosdb_sql_database.terraform.resource_group_name
  account_name          = azurerm_cosmosdb_sql_database.terraform.account_name
  database_name         = azurerm_cosmosdb_sql_database.terraform.name
  partition_key_path    = "/resourceProvider"
  partition_key_version = 2
  unique_key {
    paths = ["/path", "/operation", "/prop"]
  }
}
