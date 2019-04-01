import React from 'react'

import { Card, Elevation, Button } from '@blueprintjs/core'
import { DISCUSS } from '../constants'

const ListItem= ({ item, index, onPinClick }) => (
  <Card
    className='list-item-card'
    elevation={Elevation.TWO}
    key={item.id}
  >
    <div className='card-title'>
      <div className='card-discuss-container'>
        <a
          href={item.url}
          ping={item.encryptedURL}
          target='_blank'
        >
          {item.title}
          <span className='domain-text'>{` (${item.domain})`}</span>
          <span
            className='discuss-link'
            onClick={(e) => {
              e.stopPropagation()
              window.open(item.discussLink, '_blank')
            }}
          >
            {DISCUSS}
          </span>
        </a>
      </div>
      <div className='action-group'>
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
