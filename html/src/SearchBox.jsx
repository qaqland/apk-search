import { SearchBox } from 'react-instantsearch'

const MySearchBox = ({ classNames }) => {
  return (
    <SearchBox
      placeholder="Search for packages"
      autoFocus
      classNames={{
        root: classNames,
        form: 'flex justify-between items-center rounded-md focus-within:ring-1 ring-slate-900/20 border p-1 border-gray-200 shadow focus-within:shadow-md',
        input: 'px-2 py-1 focus:outline-none grow ',
        submit: 'px-2',
        submitIcon: 'w-4 h-4',
        reset: 'hidden ',
        resetIcon: 'hidden ',
        loadingIndicator: 'hidden',
        loadingIcon: 'hidden',
      }}
    />
  )
}

export default MySearchBox
