export function detectRegistry(imageName: string): { registryName: string; repository: string } {
  if (!imageName) return { registryName: 'dockerhub', repository: imageName }

  const slash = imageName.indexOf('/')
  if (slash === -1) {
    return { registryName: 'dockerhub', repository: `library/${imageName}` }
  }

  const prefix = imageName.slice(0, slash)
  const rest = imageName.slice(slash + 1)

  if (prefix === 'ghcr.io') return { registryName: 'ghcr', repository: rest }
  if (prefix === 'quay.io') return { registryName: 'quay', repository: rest }
  if (prefix === 'gcr.io') return { registryName: 'gcr', repository: rest }

  if (!prefix.includes('.')) {
    return { registryName: 'dockerhub', repository: imageName }
  }

  return { registryName: 'dockerhub', repository: imageName }
}
