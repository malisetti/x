import React from 'react'

import { ButtonGroup, Button } from '@blueprintjs/core'

import { PIN_FILTERS } from '../constants'

const PinnedFilter = ({ handlePinnedFilterClick }) => (
  <ButtonGroup>
    {
      PIN_FILTERS.map((filter, index) => (
        <Button
          key={index}
          text={filter}
          onClick={() => handlePinnedFilterClick(filter)}
        />
      ))
    }
  </ButtonGroup>
)

export default PinnedFilter