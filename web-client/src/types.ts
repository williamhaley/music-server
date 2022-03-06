import { Dispatch } from "react";

export interface State {
  index: number;
  currentTrack: Track | null;
  playlist: Track[];
}

export interface AnyAction {
  type: string;
  payload: any;
}

export type StoreContextType = {
  state: State;
  dispatch: Dispatch<AnyAction>;
};

export interface Track {
  id: string;
  name: string;
  extension: string;
}

export interface Album {
  id: string;
  name: string;
  tracks: Array<Track>;
}
