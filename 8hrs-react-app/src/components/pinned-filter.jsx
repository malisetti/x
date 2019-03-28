import React from 'react'

import { ButtonGroup, Button } from '@blueprintjs/core'

import { PIN_FILTERS } from '../constants'

const PinnedFilter = ({ value, onPinFilterClick }) => (
  <ButtonGroup>
    {
      PIN_FILTERS.map((filter, index) => (
        <Button
          key={index}
          active={value === filter}
          text={filter}
          onClick={() => onPinFilterClick(filter)}
        />
      ))
    }
  </ButtonGroup>
)

export default PinnedFilter