@description('The location in which all resources should be deployed.')
param location string = resourceGroup().location

// @description('The Runtime stack of current web app')
// param linuxFxVersion string = 'DOCKER|index.docker.io/appsvc/sample-hello-world'

var appName = 'app-minatsugo-bot'
var appServicePlanName = 'go-${appName}${uniqueString(subscription().subscriptionId)}'
var appServiceManagedIdentityName = 'id-${appName}'
var acrName = 'acr${uniqueString(subscription().subscriptionId)}'


// Managed Identity for App Service
resource appServiceManagedIdentity 'Microsoft.ManagedIdentity/userAssignedIdentities@2023-01-31' = {
  name: appServiceManagedIdentityName
  location: location
}

//App service plan
resource appServicePlan 'Microsoft.Web/serverfarms@2023-01-01' = {
  name: appServicePlanName
  location: location
  sku: {
    name: 'F1'
    capacity:1
  }
  properties: {
    zoneRedundant: false
    reserved: true
  }
  kind: 'linux'
}

// Web App
resource webApp 'Microsoft.Web/sites@2023-01-01' = {
  name: appName
  location: location
  // identity: {
  //   type: 'UserAssigned'
  //   userAssignedIdentities: {
  //     '${appServiceManagedIdentity.id}': {}
  //   }
  // }
  properties: {
    serverFarmId: appServicePlan.id
    siteConfig: {
      // linuxFxVersion: 'DOCKER|index.docker.io/appsvc/sample-hello-world'
      linuxFxVersion: 'DOCKER|index.docker.io/appsvc/sample-hello-world'
    }
  }
}

resource acrResource 'Microsoft.ContainerRegistry/registries@2023-01-01-preview' = {
  name: acrName
  location: location
  sku: {
    name: 'Basic'
  }
  // identity: {
  //   type: 'UserAssigned'
  //   userAssignedIdentities: {
  //     '${appServiceManagedIdentity.id}': {}
  //   }
  // }
  properties: {
    adminUserEnabled: true
  }
}
