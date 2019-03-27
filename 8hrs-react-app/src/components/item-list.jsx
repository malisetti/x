import React from 'react'

import { DISCUSS, PIN, UNPIN } from '../constants'

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

const ItemList = ({ items, handlePinClick }) => (
  items.map((item, index) => (
    <ListItem key={item.id} item={item} index={index} onPinClick={handlePinClick} />
  ))
)

export default ItemList
