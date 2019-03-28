import { TIME_FRAMES, PIN_FILTERS } from '../constants'

export const SET_PINNED_ITEMS = 'data/SET_PINNED_ITEMS'
export const GET_DATA = 'data/GET_DATA'
export const REVERSE_ITEMS = 'data/REVERSE_ITEMS'
export const PIN_ITEM = 'data/PIN_ITEM'
export const UNPIN_ITEM = 'data/UNPIN_ITEM'
export const CHANGE_PIN_FILTER = 'data/CHANGE_PIN_FILTER'
export const SET_LOADING = 'data/SET_LOADING'

const initialState = {
  timeFrame: TIME_FRAMES[0],
  pinnedItems: [],
  items: [],
  isReversed: false,
  pinFilter: PIN_FILTERS[0],
  isLoading: true,
}

export default (state = initialState, action) => {
  switch (action.type) {
    case SET_PINNED_ITEMS:
      return {
        ...state,
        pinnedItems: action.payload.pinnedItems,
      }
    case GET_DATA:
      return {
        ...state,
        items: action.payload.items,
        timeFrame: action.payload.timeFrame,
      }
    case REVERSE_ITEMS:
      return {
        ...state,
        items: [...state.items].reverse(),
        isReversed: !state.isReversed,
      }
    case PIN_ITEM: {
      return {
        ...state,
        pinnedItems: [action.payload.pinnedItem, ...state.pinnedItems],
      }
    }
    case UNPIN_ITEM: {
      return {
        ...state,
        pinnedItems: state.pinnedItems.filter(item => item.id !== action.payload.unpinnedId),
      }
    }
    case CHANGE_PIN_FILTER: {
      return {
        ...state,
        pinFilter: action.payload.pinFilter,
      }
    }
    case SET_LOADING: {
      return {
        ...state,
        isLoading: action.payload.isLoading,
      }
    }
    default:
      return state
  }
}
