import { useReducer, createContext } from "react";
import { AnyAction, State, StoreContextType } from "./types";

const initialState: State = {
  index: 0,
  currentTrack: null,
  playlist: [],
};

const reducer = (state: State, action: AnyAction) => {
  switch (action.type) {
    case 'QUEUE_AND_PLAY':
      console.log('tracks queued:', action.payload);
      return { ...state, currentTrack: action.payload[0], index: 0, playlist: action.payload };
    case 'NEXT_TRACK':
      let nextIndex = state.index >= state.playlist.length - 1 ? 0 : state.index + 1;
      return { ...state, index: nextIndex, currentTrack: state.playlist[nextIndex] };
    case 'PREVIOUS_TRACK':
      let previousIndex = state.index === 0 ? state.playlist.length - 1 : state.index - 1;
      return { ...state, index: previousIndex, currentTrack: state.playlist[previousIndex] };
    default:
      throw new Error(`invalid action: ${action.type}`);
  }
};

export const StoreContext = createContext<StoreContextType>({
  state: {
    index: 0,
    currentTrack: null,
    playlist: []
  },
  dispatch: () => { },
});

export interface StoreProps {
  children: JSX.Element;
}

const Store = ({ children }: StoreProps) => {
  const [state, dispatch] = useReducer(reducer, initialState);

  return (
    <StoreContext.Provider value={{ state, dispatch }}>
      {children}
    </StoreContext.Provider>
  );
};

export default Store;