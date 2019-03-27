import React from 'react'

import { ButtonGroup, Button } from '@blueprintjs/core'

import { TIME_FRAMES, HRS } from '../constants'

const TimeFrames = ({ handleTimeFrameClick }) => (
  <ButtonGroup className='time-frame-group'>
    {
      TIME_FRAMES.map((timeFrame, idx) => (
        <Button
          className='time-frame-button'
          text={`${timeFrame}${HRS}`}
          onClick={() => handleTimeFrameClick(timeFrame)}
        />
      ))
    }
  </ButtonGroup>
)

export default TimeFrames