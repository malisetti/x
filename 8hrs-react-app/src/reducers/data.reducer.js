export const GET_DATA = 'data/GET_DATA'
export const REVERSE_ITEMS = "data/REVERSE_ITEMS"
export const PIN_ITEM = "data/PIN_ITEM"
export const UNPIN_ITEM = "data/UNPIN_ITEM"

const initialState = {
  pinnedItemsIds: [],
  items: {},
}

const itemsReducer = (acc, item) => ({
  ...acc,
  [item.id]: item,
})

export default (state = initialState, action) => {
  switch (action.type) {
    case GET_DATA:
      return {
        ...state,
        items: action.payload.items.reduce(itemsReducer, {}),
      }
    case REVERSE_ITEMS:
      return {
        ...state,
        items: Object.values(state.items).reverse().reduce(itemsReducer, {}),
      }
    case PIN_ITEM: {
      return {
        ...state,
        pinnedItemsIds: [action.payload.pinnedId, ...pinnedItemsIds],
      }
    }
    case UNPIN_ITEM: {
      return {
        ...state,
        pinnedItemsIds: state.pinnedItemsIds.filter(id => id !== action.payload.unPinnedId),
      }
    }

    default:
      return state
  }
}
