import React from 'react'
import { bindActionCreators } from 'redux'
import { connect } from 'react-redux'

import { Switch, ProgressBar } from '@blueprintjs/core'

import * as DataActions from '../../actions/data.actions'
import { TITLE, REVERSE, PIN_FILTERS } from '../../constants'

import ItemList from '../../components/item-list'
import TimeFrames from '../../components/time-frames'
import PinnedFilter from '../../components/pinned-filter'

class Home extends React.Component {

  componentDidMount() {
    this.props.getData()
    this.props.getPinnedItems()
  }

  handleTimeFrameClick = (timeFrame) => {
    this.props.getData(timeFrame)
  }

  handleReverseClick = () => {
    this.props.reverseItems()
  }

  handlePinClick = (item, isPinned) => {
    if (isPinned) {
      this.props.unpinItem(item.id)
      return
    }
    this.props.pinItem(item)
  }

  handlePinFilterClick = (value) => {
    this.props.changePinFilter(value)
  }

  render() {
    const { timeFrame, isReversed, allItems, pinFilter, isLoading } = this.props
    const { items, pinnedItems } = allItems
    const itemsToDisplay = pinFilter === PIN_FILTERS[0] ? items : pinnedItems
    return (
      <div className='bp3-dark'>
        <h1>{TITLE}</h1>
        <div className='filter-container'>
          <TimeFrames
            value={timeFrame}
            onTimeFrameClick={this.handleTimeFrameClick} />
          <PinnedFilter
            value={pinFilter}
            onPinFilterClick={this.handlePinFilterClick}/>
        </div>
        {
          isLoading && <ProgressBar className='progress-bar'/>
        }
        {
          pinFilter === PIN_FILTERS[0]
          && <Switch checked={isReversed} label={REVERSE} onChange={this.handleReverseClick} />
        }
        {
          !!itemsToDisplay.length && <ItemList items={itemsToDisplay} handlePinClick={this.handlePinClick} />
        }
        
      </div>
    )
  }
}

const arrangeItems = (items, pinnedItems) => {
  const pinnedItemsIds = pinnedItems.map(item => item.id)
  const otherItems = items.map(item => {
    if (!pinnedItemsIds.includes(item.id)) return { ...item, isPinned: false }
    return { ...item, isPinned: true }
  })
  const formattedPinnedItems = pinnedItems.map(item => ({ ...item, isPinned: true }))
  return {
    pinnedItems: [...formattedPinnedItems],
    items: [...otherItems],
  }
}

const mapStateToProps = ({ data }) => ({
  allItems: arrangeItems(data.items, data.pinnedItems),
  timeFrame: data.timeFrame,
  isReversed: data.isReversed,
  pinFilter: data.pinFilter,
  isLoading: data.isLoading,
})

const mapDispatchToProps = dispatch =>
  bindActionCreators(DataActions, dispatch)

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(Home)
