import axios from "axios";

import { GET_DATA, REVERSE_ITEMS, PIN_ITEM, UNPIN_ITEM, CHANGE_PIN_FILTER, SET_LOADING, SET_PINNED_ITEMS } from '../reducers/data.reducer.js'
import { BASE_URL, TIME_FRAMES, PINNED_ITEMS_KEY } from '../constants'

export const setLoading = (isLoading) => ({
  type: SET_LOADING,
  payload: {
    isLoading,
  }
})

const setPinnedItems = (pinnedItems) => ({
  type: SET_PINNED_ITEMS,
  payload: {
    pinnedItems,
  }
})

export const getPinnedItems = () => {
  return dispatch => {
    const pinnedItems = JSON.parse(localStorage.getItem(PINNED_ITEMS_KEY)) || []
    dispatch(setPinnedItems(pinnedItems))
  }
}

export const getData = (timeFrame = TIME_FRAMES[0]) => {
  return dispatch => {
    const params = { headers: { 'Content-Type': 'application/json' }, data: {}, params: { t: timeFrame } }
    // dispatch(setLoading(true))
    axios.get(BASE_URL, params)
      .then((response) => {
        dispatch({
          type: GET_DATA,
          payload: {
            items: response.data,
            timeFrame,
          },
        })
        dispatch(setLoading(false))
      })
      .catch((error) => {
        dispatch(setLoading(false))
      });
  }
}

export const reverseItems = () => ({
  type: REVERSE_ITEMS,
})

export const pinItem = (item) => {
  return dispatch => {
    const pinnedItems = JSON.parse(localStorage.getItem(PINNED_ITEMS_KEY)) || []
    pinnedItems.unshift(item)
    localStorage.setItem(PINNED_ITEMS_KEY, JSON.stringify(pinnedItems))
    dispatch(sendPinItem(item))
  }
}

const sendPinItem = (item) => ({
  type: PIN_ITEM,
  payload: {
    pinnedItem: item,
  },
})

export const unpinItem = (id) => {
  return dispatch => {
    const pinnedItems = JSON.parse(localStorage.getItem(PINNED_ITEMS_KEY)) || []
    const modifiedPinnedItems = pinnedItems.filter(item => item.id !== id)
    localStorage.setItem(PINNED_ITEMS_KEY, JSON.stringify(modifiedPinnedItems))
    dispatch(sendUnpinItem(id))
  }
}

const sendUnpinItem = (id) => ({
  type: UNPIN_ITEM,
  payload: {
    unpinnedId: id,
  },
})

export const changePinFilter = (pinFilter) => ({
  type: CHANGE_PIN_FILTER,
  payload: {
    pinFilter,
  }
})
