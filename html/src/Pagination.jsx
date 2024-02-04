import { Pagination } from 'react-instantsearch'

// TODO: click and scroll to top

const MyPagination = ({ classNames }) => {
  return (
    <Pagination
      showFirst={false}
      showPrevious={false}
      showNext={false}
      showLast={false}
      classNames={{
        root: classNames,
        list: 'flex justify-center space-x-2 items-baseline',
        item: 'w-8 h-7  rounded hover:shadow shadow-sm border',
        link: ' inline-block size-full text-center align-baseline ',
        selectedItem: 'text-gray-900/50 bg-gray-100 *:cursor-not-allowed',
        disabledItem: 'text-gray-900 ',
      }}
    />
  )
}

export default MyPagination
