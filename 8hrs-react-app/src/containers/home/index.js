import React from 'react'
import { bindActionCreators } from 'redux'
import { connect } from 'react-redux'

import * as DataActions from '../../actions/data.actions'
import { TITLE, REVERSE } from '../../constants'

// const items = require('../../items.json')
import ClickableLink from '../../components/clickable-link'
import ItemList from '../../components/item-list'
import TimeFrames from '../../components/time-frames'
import PinnedFilter from '../../components/pinned-filter'

class Home extends React.Component {

  componentDidMount() {
    this.props.getData()
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

  render() {
    const { timeFrame, allItems } = this.props
    const { items, pinnedItems } = allItems
    return (
      <div className='bp3-dark'>
        <h1>{TITLE}</h1>
        <div className='filter-container'>
          <TimeFrames
            value={timeFrame}
            handleTimeFrameClick={this.handleTimeFrameClick} />
          <PinnedFilter />
        </div>
        {
          !!pinnedItems.length &&
          <React.Fragment>
            <ItemList items={pinnedItems} handlePinClick={this.handlePinClick} />
            <hr className='separator'/>
          </React.Fragment>
        }
        <ClickableLink
          className='reverse-link'
          id='reverse-items'
          onClick={this.handleReverseClick}
        >
          {REVERSE}
        </ClickableLink>
        <ItemList items={items} handlePinClick={this.handlePinClick} />
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
})

const mapDispatchToProps = dispatch =>
  bindActionCreators(DataActions, dispatch)

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(Home)
