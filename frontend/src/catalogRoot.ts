import { catalogPath, taxonomySlug } from './routes'
import type { CatalogRoot } from './api'

export function catalogRootViewModel(root: CatalogRoot) {
  return root.groups.map(group => {
    const rootPath = [taxonomySlug(group.root.label)]
    return {
      ...group,
      headingHref: catalogPath({ taxonomyDef: group.taxonomyDefId, path: rootPath }),
      children: group.children.map(child => ({
        ...child,
        href: catalogPath({ taxonomyDef: group.taxonomyDefId, path: [...rootPath, taxonomySlug(child.label)] }),
      })),
    }
  })
}
