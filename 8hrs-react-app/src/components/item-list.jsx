import React from 'react'

import { Card, Elevation, Button } from '@blueprintjs/core'

const ListItem= ({ item, index, onPinClick }) => (
  <Card
    className='list-item-card'
    elevation={Elevation.TWO}
    key={item.id}
    onClick={() => window.open(item.url, '_blank')}
  >
    <div className='card-title'>
      <h4>
        <a
          href='#'
          onClick={(e) => e.preventDefault()}
        >
          {item.title}
          <span className='domain-text'>{` (${item.domain})`}</span>
        </a>
      </h4>
      <div className='action-group'>
        <Button
          className='action-button'
          icon="comment"
          onClick={(e) => {
            e.stopPropagation()
            window.open(item.discussLink, '_blank')
          }}
        />
        <Button
          className='action-button'
          icon="pin"
          active={item.isPinned}
          onClick={(e) => {
            e.stopPropagation()
            onPinClick(item, item.isPinned)
          }}
        />
      </div>
    </div>
</Card>
)

const ItemList = ({ items, handlePinClick }) => (
  items.map((item, index) => (
    <ListItem key={item.id} item={item} index={index} onPinClick={handlePinClick} />
  ))
)

export default ItemList
