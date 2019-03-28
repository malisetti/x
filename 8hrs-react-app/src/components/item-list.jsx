import React from 'react'

import { Card, Elevation, Button } from '@blueprintjs/core'

import { DISCUSS, PIN, UNPIN } from '../constants'

const ListItem1 = ({ item, index, onPinClick }) => (
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

const ListItem= ({ item, index, onPinClick }) => (
  <Card
    interactive={true}
    elevation={Elevation.TWO}
    key={item.id}
  >
    <h5><a href="#">{item.title}</a></h5>
    <p>{item.description}</p>
    <Button>Pin</Button>
</Card>
)

const ItemList = ({ items, handlePinClick }) => (
  items.map((item, index) => (
    <ListItem key={item.id} item={item} index={index} onPinClick={handlePinClick} />
  ))
)

export default ItemList
