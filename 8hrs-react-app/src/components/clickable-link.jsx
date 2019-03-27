import React from 'react'

const ClickableLink = ({ className, id, onClick, children }) => (
  <p className={className}>
    <a
      key={id}
      href="#"
      onClick={(e) => {
        e.preventDefault()
        onClick(id)
      }}
    >
      {children}
    </a>
  </p>
)

export default ClickableLink