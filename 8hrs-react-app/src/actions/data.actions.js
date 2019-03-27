import axios from "axios";

import { GET_DATA, REVERSE_ITEMS, PIN_ITEM, UNPIN_ITEM } from '../reducers/data.reducer.js'
import { BASE_URL, TIME_FRAMES } from '../constants'

export const getData = (timeFrame = TIME_FRAMES[0]) => {
  return dispatch => {
    axios
      .get(BASE_URL, { headers: { 'Content-Type': 'application/json' }, data: {}, params: { t: timeFrame } })
      .then((response) => {
        dispatch({
          type: GET_DATA,
          payload: {
            items: response.data,
          },
        })
      })
      .catch((error) => {});
  }
}

export const reverseItems = () => ({
  type: REVERSE_ITEMS,
})

export const pinItem = (item) => ({
  type: PIN_ITEM,
  payload: {
    pinnedItem: item,
  },
})

export const unpinItem = (id) => ({
  type: UNPIN_ITEM,
  payload: {
    unpinnedId: id,
  },
})
