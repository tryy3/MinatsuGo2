@description('The location in which all resources should be deployed.')
param location string = resourceGroup().location

@description('The name of the resource group in which all resources should be deployed.')
param servicePrincipalObjectId string = ''

// @description('The Runtime stack of current web app')
// param linuxFxVersion string = 'DOCKER|index.docker.io/appsvc/sample-hello-world'

var appName = 'app-minatsugo-bot'
var appServicePlanName = 'go-${appName}${uniqueString(subscription().subscriptionId)}'
var appServiceManagedIdentityName = 'id-${appName}'
var acrName = 'acr${uniqueString(subscription().subscriptionId)}'
var keyVaultName = 'kv-${appName}'


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
  identity: {
    type: 'UserAssigned'
    userAssignedIdentities: {
      '${appServiceManagedIdentity.id}': {}
    }
  }
  kind: 'linux'
  properties: {
    serverFarmId: appServicePlan.id
    reserved: true
    siteConfig: {
      // linuxFxVersion: 'DOCKER|index.docker.io/appsvc/sample-hello-world'
      linuxFxVersion: 'DOCKER|acrawwzg43ei25ai.azurecr.io/acrawwzg43ei25ai/test:5a7fd6092dfbdbc1e53831dff5c3b22f26f548d8'
      appSettings: [
        { name: 'VAULT_URL', value: keyVault.properties.vaultUri }
      ]
    }
  }
}

// resource acrRoleAssignment 'Microsoft.Authorization/roleAssignments@2020-04-01-preview' = {
//   name: guid(acrResource.id, 'AcrPull')
//   scope: acrResource
//   properties: {
//     roleDefinitionId: resourceId('Microsoft.Authorization/roleDefinitions', '7f951dda-4ed3-4680-a7ca-43fe172d538d') // AcrPull role
//     principalId: appServiceManagedIdentity.properties.principalId
//     principalType: 'ServicePrincipal'
//   }
// }

resource acrResource 'Microsoft.ContainerRegistry/registries@2023-01-01-preview' = {
  name: acrName
  location: location
  sku: {
    name: 'Basic'
  }
  identity: {
    type: 'UserAssigned'
    userAssignedIdentities: {
      '${appServiceManagedIdentity.id}': {}
    }
  }
  properties: {
    adminUserEnabled: true
  }
}

resource keyVault 'Microsoft.KeyVault/vaults@2023-07-01' = {
  name: keyVaultName
  location: location
  properties: {
    sku: {
      family: 'A'
      name: 'standard'
    }
    tenantId: subscription().tenantId
    accessPolicies: [
      {
        tenantId: subscription().tenantId
        objectId: appServiceManagedIdentity.properties.principalId
        permissions: {
          keys: ['get', 'list']
          secrets: ['get', 'list']
          certificates: ['get', 'list']
        }
      }
      {
        tenantId: subscription().tenantId
        objectId: servicePrincipalObjectId
        permissions: {
          secrets: ['set']
        }
      }
    ]
  }
}
