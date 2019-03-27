import React from 'react'
import { bindActionCreators } from 'redux'
import { connect } from 'react-redux'

import * as DataActions from '../../actions/data.actions'
import { TITLE, DISCUSS, TIME_FRAMES, HRS, PIN, UNPIN } from '../../constants'

// const items = require('../../items.json')

const ClickableLink = ({ className, id, onClick, children }) => (
  <a
    className={className}
    key={id}
    href="#"
    onClick={(e) => {
      e.preventDefault()
      onClick(id)
    }}
  >{children}</a>
)

const ListItem = ({ item, index, onPinClick }) => (
  <div className='list-item' key={item.id}>
    <a
      href={item.url}
      target="_blank"
    >
      {`${index + 1}. ${item.title}`}
    </a>
    <a
      className='domain-link'
      href={item.url}
      target="_blank"
    >
      {`(${item.domain})`}
    </a>
    <a
      className='discuss-link'
      href={item.discussLink}
      target='_blank'
      ping={`/l/${item.encryptedDiscussLink}`}
    >
      {DISCUSS}
    </a>
    <a
      className='pinning-link'
      href="#"
      onClick={(e) => {
        e.preventDefault()
        onPinClick(item, item.isPinned)
      }}
    >
      {item.isPinned ? UNPIN : PIN}
    </a>
  </div>
)

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
    const { items, pinnedItems } = this.props.allItems
    return(
      <div>
        <h1>{TITLE}</h1>
        <div className='time-frames'>
          {
            TIME_FRAMES.map((timeFrame, idx) => (
              <ClickableLink
                className='time-frame-link'
                key={idx}
                onClick={() => this.handleTimeFrameClick(timeFrame)}
              >
                {`${timeFrame}${HRS}`}
              </ClickableLink>
            ))
          }
        </div>
        {
          !!pinnedItems.length &&
          <React.Fragment>
            { pinnedItems.map((item, index) => <ListItem key={item.id} item={item} index={index} onPinClick={this.handlePinClick} />) }
            <hr className='separator'/>
          </React.Fragment>
        }
        <ClickableLink
          className='time-frame-link'
          id='reverse-items'
          onClick={this.handleReverseClick}
        >
          reverse
        </ClickableLink>
        {
          items.map((item, index) => <ListItem key={item.id} item={item} index={index} onPinClick={this.handlePinClick} />)
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
  allItems: arrangeItems(data.items, data.pinnedItems) 
})

const mapDispatchToProps = dispatch =>
  bindActionCreators(DataActions, dispatch)

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(Home)
