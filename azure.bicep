@description('The location in which all resources should be deployed.')
param location string = resourceGroup().location

var appName = 'app-minatsugo-bot'
var appServicePlanName = 'asp-${appName}${uniqueString(subscription().subscriptionId)}'
var appServiceManagedIdentityName = 'id-${appName}'


// Managed Identity for App Service
resource appServiceManagedIdentity 'Microsoft.ManagedIdentity/userAssignedIdentities@2023-01-31' = {
  name: appServiceManagedIdentityName
  location: location
}

//App service plan
resource appServicePlan 'Microsoft.Web/serverfarms@2022-09-01' = {
  name: appServicePlanName
  location: location
  sku: {
    name: 'F1'
    capacity:1
  }
  properties: {
    zoneRedundant: false
  }
  kind: 'app'
}

// Web App
resource webApp 'Microsoft.Web/sites@2022-09-01' = {
  name: appName
  location: location
  kind: 'app'
  identity: {
    type: 'UserAssigned'
    userAssignedIdentities: {
      '${appServiceManagedIdentity.id}': {}
    }
  }
  properties: {
    serverFarmId: appServicePlan.id
    httpsOnly: false
    hostNamesDisabled: false
    siteConfig: {
      vnetRouteAllEnabled: false
      http20Enabled: true
      publicNetworkAccess: 'Enabled'
      alwaysOn: true
    }
  }
}

// App Settings
resource appsettings 'Microsoft.Web/sites/config@2022-09-01' = {
  name: 'appsettings'
  parent: webApp
  properties: {
  }
}
