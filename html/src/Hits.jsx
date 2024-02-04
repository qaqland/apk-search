import { useHits } from 'react-instantsearch'
import { useState, useEffect, useId } from 'react'
import { MeiliSearch } from 'meilisearch'
import { Host, Key } from './Key.jsx'

const mljsClient = new MeiliSearch({
  host: Host,
  apiKey: Key,
})

const Hit = ({ hit, packageID, indexName }) => {
  const [details, setDetails] = useState(null)

  useEffect(() => {
    if (packageID != hit.objectID) {
      return
    }
    if (details != null) {
      return
    }
    mljsClient
      .index(indexName)
      .getDocument(packageID)
      .then((res) => {
        setDetails(res)
      })
      .catch((err) => {
        console.log(err)
      })
  }, [packageID])

  const lableId = useId()
  const regx = /(edge|v\d_\d+)_(.+)/
  const match = indexName.match(regx)
  const branch = match[1].replace('_', '.')
  const arch = match[2]
  const officalpackage = `pkgs.alpinelinux.org/package/${branch}/${hit.repository}/${arch}/${hit.origin}`
  const gitbranch = branch == 'edge' ? 'master' : branch.substr(1) + '-stable'
  const cgitlink = `https://git.alpinelinux.org/aports/tree/${hit.repository}/${hit.origin}?h=${gitbranch}`

  return (
    <label htmlFor={lableId}>
      <div className="mt-4 rounded-md border border-gray-200 shadow ring-slate-900/10 hover:bg-gray-50/20 hover:shadow-md hover:ring-1">
        <div className="flex items-baseline space-x-2 rounded-t-md bg-gray-50 px-2 py-1 text-sm text-gray-600">
          {hit.origin == hit.package ? null : (
            <>
              <span className="text-xs font-bold text-gray-600/70">FROM</span>
              <span className="">{hit.origin}</span>
            </>
          )}
          <span className="text-xs font-bold text-gray-600/70">IN</span>
          <span className="">{hit.repository}</span>

          <span className="!ml-auto text-xs font-bold text-gray-600/70">
            {new Date(hit.build_time * 1000).toLocaleDateString()}
          </span>
        </div>
        <div className="flex flex-wrap items-baseline justify-between border-t border-gray-900/10 px-2 pt-2 ">
          <span className="mr-4 select-all text-gray-900">{hit.package}</span>
          <span className="text-sm text-gray-600 ">v{hit.version}</span>
        </div>
        <div className="mx-2 mb-3 mt-0.5 text-pretty pl-1 text-sm tracking-tight text-gray-900 ">
          {hit.description}
        </div>
      </div>
      <input type="radio" name="hit" id={lableId} className="peer hidden" />
      <div className="mt-2 hidden border-gray-900/10 px-2 text-sm peer-checked:block">
        {details ? (
          <div className="hover:*:ring-0.5 flex flex-row flex-wrap ring-slate-800/10 *:m-1 *:rounded-md *:border *:border-gray-300/60  *:shadow-sm *:ring-slate-800/10 hover:*:shadow">
            <span className="flex items-baseline">
              <span className="rounded-l-md border-r bg-gray-100 px-2 py-1  text-gray-600/90">
                upstream
              </span>
              <a
                href={details.project}
                target="_blank"
                className="truncate px-2 py-1 "
              >
                {details.project}
              </a>
            </span>
            <span className="flex items-baseline">
              <span className="rounded-l-md border-r bg-gray-100 px-2 py-1  text-gray-600/90">
                license
              </span>
              <span className="truncate px-2 py-0.5 ">{details.license}</span>
            </span>
            <span className="flex items-baseline">
              <span className="rounded-l-md border-r bg-gray-100 px-2 py-1  text-gray-600/90">
                size
              </span>
              <span className="truncate px-2 py-0.5">
                {details.file_size}/{details.installed_size}
              </span>
            </span>
            <span className="flex items-baseline">
              <span className="rounded-l-md border-r bg-gray-100 px-2 py-1  text-gray-600/90">
                package
              </span>
              <a
                href={'https://' + officalpackage}
                target="_blank"
                className="truncate px-2 py-0.5 "
              >
                {officalpackage}
              </a>
            </span>
            <span className="flex items-baseline">
              <span className="rounded-l-md border-r bg-gray-100 px-2 py-1  text-gray-600/90">
                maintainer
              </span>
              <a
                href={cgitlink}
                target="_blank"
                className="truncate px-2 py-0.5"
              >
                {details.maintainer.name}
              </a>
            </span>
            <span className="flex items-baseline">
              <span className="rounded-l-md border-r bg-gray-100 px-2 py-1  text-gray-600/90">
                commit
              </span>
              <a
                href={
                  'https://gitlab.alpinelinux.org/alpine/aports/-/commit/' +
                  details.commit
                }
                target="_blank"
                className="truncate px-2 py-0.5"
              >
                {details.commit}
              </a>
            </span>
            {details.depends?.map((d) => (
              <span className="flex items-baseline" key={d}>
                <span className="rounded-l-md border-r bg-gray-100 px-2 py-1  text-gray-600/90">
                  require
                </span>
                <span className="truncate px-2 py-0.5">{d}</span>
              </span>
            ))}
            {details.provides?.map((d) => (
              <span className="flex items-baseline" key={d}>
                <span className="rounded-l-md border-r bg-gray-100 px-2 py-1  text-gray-600/90">
                  provide
                </span>
                <span className="truncate px-2 py-0.5">{d}</span>
              </span>
            ))}
          </div>
        ) : (
          <div>loading...</div>
        )}
      </div>
    </label>
  )
}

const MyHits = (props) => {
  const { hits } = useHits(props)
  const { classNames, indexName, packageID, setPackage } = props

  return (
    <div className={classNames}>
      <ol>
        {hits.map((hit) => (
          <li
            key={hit.objectID}
            onClick={() => setPackage(hit.objectID)}
            onAuxClick={() => setPackage(hit.objectID)}
          >
            <Hit hit={hit} packageID={packageID} indexName={indexName} />
          </li>
        ))}
      </ol>
    </div>
  )
}

export default MyHits
