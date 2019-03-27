export const GET_DATA = 'data/GET_DATA'
export const REVERSE_ITEMS = "data/REVERSE_ITEMS"
export const PIN_ITEM = "data/PIN_ITEM"
export const UNPIN_ITEM = "data/UNPIN_ITEM"

const initialState = {
  pinnedItems: [],
  items: [],
}

export default (state = initialState, action) => {
  switch (action.type) {
    case GET_DATA:
      return {
        ...state,
        items: action.payload.items,
      }
    case REVERSE_ITEMS:
      return {
        ...state,
        items: [...state.items].reverse(),
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

    default:
      return state
  }
}
