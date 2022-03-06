import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { Album } from "./types";
import { useAuth } from "./useAuth";

const getAlbums = async (token: string) => {
  const res = await fetch('/api/albums', {
    headers: {
      'Access-Token': token,
    },
  });

  const json = await res.json();

  return json.data;
};

const Albums = () => {
  const [albums, setAlbums] = useState([]);
  const { authState } = useAuth();

  useEffect(() => {
    let didAbort = false;

    (async () => {
      const latestAlbums = await getAlbums(authState.token ?? '');

      if (didAbort) {
        return;
      }

      setAlbums(latestAlbums);
    })();

    return () => { didAbort = true };
  }, [authState.token]);

  return (
    <>
      <h1>Albums</h1>

      <ul>
        {albums.map((album: Album) => {
          return (
            <li key={album.id}>
              <Link to={`/albums/${album.id}`}>{album.name}</Link>
            </li>
          );
        })}
      </ul>
    </>
  );
};

export default Albums;