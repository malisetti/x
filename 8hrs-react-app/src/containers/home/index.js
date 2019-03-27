import React from 'react'
import { bindActionCreators } from 'redux'
import { connect } from 'react-redux'

import * as DataActions from '../../actions/data.actions'
import { TIME_FRAMES, HRS } from '../../constants'

const TITLE = 'Tech & News'
const DISCUSS = 'discuss'

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
      className='discuss-link'
      href={item.discussLink}
      target='_blank'
      ping={`/l/${item.encryptedDiscussLink}`}
    >
      {DISCUSS}
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

  handlePinClick = (id, action) => {
    this.props.
  }

  render() {
    const { items } = this.props
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
        <ClickableLink
          className='time-frame-link'
          id='reverse-items'
          onClick={this.handleReverseClick}
        >
          reverse
        </ClickableLink>
        {
          items.map((item, index) => <ListItem item={item} index={index} onPinClick={this.handlePinning} />)
        }
      </div>
    )
  }
}

const arrangeItems = (items, pinnedItemsIds) => {

}

const mapStateToProps = ({ data }) => ({
  items: arrangeItems(data.items, data.pinnedItemsIds) 
})

const mapDispatchToProps = dispatch =>
  bindActionCreators(DataActions, dispatch)

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(Home)
