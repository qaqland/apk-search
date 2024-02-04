import { useStats } from 'react-instantsearch'

const MyStats = ({ classNames }) => {
  const { nbHits, processingTimeMS } = useStats()

  return (
    <span className={classNames}>
      hit {nbHits} {'in '}
      {processingTimeMS}ms
    </span>
  )
}

export default MyStats
