import { useContext, useEffect, useState } from "react";
import { StoreContext } from "./Store";
import { Album } from "./types";
import styles from './AlbumDetails.module.scss';
import { useAuth } from "./useAuth";
import { dummyPlay } from "./audioService";

const getAlbum = async (id: string, token: string) => {
  const res = await fetch(`/api/albums/${id}`, {
    headers: {
      'Access-Token': token,
    },
  });

  const json = await res.json();

  return json.data;
};

export interface AlbumDetailsProps {
  id: string;
};

const AlbumDetails = ({ id }: AlbumDetailsProps): JSX.Element => {
  const [album, setAlbum] = useState<Album | null>(null);
  const { authState } = useAuth();
  const { dispatch } = useContext(StoreContext);

  useEffect(() => {
    let didAbort = false;

    (async () => {
      const albumDetails = await getAlbum(id, authState.token ?? "");

      if (didAbort) {
        return;
      }

      setAlbum(albumDetails);
    })();

    return () => { didAbort = true };
  }, [id, authState.token]);

  if (!album) {
    return <div>...</div>;
  }

  return (
    <div className={styles.album}>
      <h1>{album.name}</h1>
      <ul>
        {album.tracks.map((track, index) => {
          return (
            <li key={track.id} onClick={() => {
              dummyPlay().then(() => {
                dispatch({ type: 'QUEUE_AND_PLAY', payload: album.tracks.slice(index) });
              });
            }}>{track.name}</li>
          );
        })}
      </ul>
    </div>
  );
};

export default AlbumDetails;