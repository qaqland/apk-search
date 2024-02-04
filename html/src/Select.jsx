import { useState, useEffect } from 'react'

const MySelect = ({ classNames, indexName, setIndex, setPackage }) => {
  const [indexList, setIndexList] = useState([])

  useEffect(() => {
    fetchOptions()
  }, [])

  const fetchOptions = async () => {
    try {
      const response = await fetch('/indexes.json')
      const data = await response.json()
      setIndexList(data)
    } catch (error) {
      console.error('Error fetching options:', error)
    }
  }

  if (indexList.length === 0) {
    return null
  }

  const options = indexList.map((index, i) => {
    let indexUid = index.branch.replaceAll('.', '_') + '_' + index.arch
    return (
      <option key={i} value={indexUid}>
        {index.branch} {index.arch}
      </option>
    )
  })

  const handleChange = (event) => {
    setIndex(event.target.value)
    setPackage('') // why?
  }

  return (
    <select className={classNames} value={indexName} onChange={handleChange}>
      {options}
    </select>
  )
}

export default MySelect
