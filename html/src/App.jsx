import { useState, useEffect } from 'react'
import { InstantSearch } from 'react-instantsearch'
import { instantMeiliSearch } from '@meilisearch/instant-meilisearch'
import MySearchBox from './SearchBox.jsx'
import MyStats from './Stats.jsx'
import MyHits from './Hits.jsx'
import MyPagination from './Pagination.jsx'
import MySelect from './Select.jsx'
import { Host, Key } from './Key.jsx'

const searchClient = instantMeiliSearch(Host, Key, { primaryKey: 'id' })

let lastIndex = localStorage.getItem('indexName') ||'edge_x86_64'

const App = () => {
  const [indexName, setIndex] = useState(lastIndex)
  const [packageID, setPackage] = useState('')

  useEffect(() => {
    localStorage.setItem('indexName', indexName)
  }, [indexName])
  // + ":build_time:desc"
  return (
    <InstantSearch searchClient={searchClient} indexName={indexName}>
      <div className="mx-auto mt-8 flex max-w-screen-md flex-col p-1 font-mono text-gray-900">
        <MySearchBox classNames="py-4 " />
        <div className="flex items-baseline justify-between pl-2 text-sm">
          <MyStats classNames="p-1" />
          <MySelect
            classNames="p-1 pl-2 outline-0 bg-gray-50 shadow-sm focus:bg-gray-100 rounded border border-gray-200 text-gray-700 text-sm"
            indexName={indexName}
            setIndex={setIndex}
            setPackage={setPackage}
          />
        </div>
        <MyHits
          classNames="bg-red"
          indexName={indexName}
          packageID={packageID}
          setPackage={setPackage}
        />
        <MyPagination classNames="py-8" />
      </div>
    </InstantSearch>
  )
}

export default App
