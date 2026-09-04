import { catalogPath, taxonomySlug } from './routes'
import type { CatalogRoot } from './api'

export function catalogRootViewModel(root: CatalogRoot) {
  return root.groups.map(group => {
    return {
      ...group,
      headingHref: catalogPath({ taxonomyDef: group.taxonomyDefId }),
      children: group.children.map(child => ({
        ...child,
        href: catalogPath({ taxonomyDef: group.taxonomyDefId, path: [taxonomySlug(child.label)] }),
      })),
    }
  })
}
