import React from 'react'

import { ButtonGroup, Button } from '@blueprintjs/core'

import { TIME_FRAMES, HRS } from '../constants'

const TimeFrames = ({ value, handleTimeFrameClick }) => (
  <ButtonGroup className='time-frame-group'>
    {
      TIME_FRAMES.map((timeFrame, index) => (
        <Button
          className='time-frame-button'
          key={index}
          active={value === timeFrame}
          text={`${timeFrame}${HRS}`}
          onClick={() => handleTimeFrameClick(timeFrame)}
        />
      ))
    }
  </ButtonGroup>
)

export default TimeFrames