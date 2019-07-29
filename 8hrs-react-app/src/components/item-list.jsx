import React from 'react'

import { Card, Elevation, Button } from '@blueprintjs/core'
import { DISCUSS } from '../constants'

const ListItem = ({ item, index, onPinClick }) => {
  const WrapperElement = item.deleted ? "del" : item.flagged ? "s": "React.Fragment";
  return (
    <Card
      className='list-item-card'
      elevation={Elevation.TWO}
      key={item.id}
    >
      <div className='card-title'>
        <div className='card-discuss-container'>
          <WrapperElement>
            <a
              href={item.url}
              ping={`/l/${item.encryptedURL}`}
              target='_blank'
            >
              {item.title}
              <span className='domain-text'>{` (${item.domain})`}</span>
            </a>
            <a
              className='discuss-link'
              href={item.discussLink}
              target='_blank'
            >
              <i>{DISCUSS}</i>
            </a>
          </WrapperElement>
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
}

const ItemList = ({ items, handlePinClick }) => (
  items.map((item, index) => (
    <ListItem key={item.id} item={item} index={index} onPinClick={handlePinClick} />
  ))
)

export default ItemList
