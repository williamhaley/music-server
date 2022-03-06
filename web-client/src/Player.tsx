import { useContext, useEffect, useState } from 'react';
import { StoreContext } from './Store';
import styles from './Player.module.scss';
import classnames from 'classnames';
import { useAuth } from './useAuth';
import PlayerTime from './PlayerTime';
import PlayerControls from './PlayerControls';
import { audio, load, pause, play } from './audioService';

const Player = () => {
  const { state, dispatch } = useContext(StoreContext);
  const [isPlaying, setIsPlaying] = useState(false);
  const [playedGarbage, setPlayedGarbage] = useState(false);
  const { authState } = useAuth();

  useEffect(() => {
    if (!state.currentTrack) {
      return;
    }

    console.log('track has changed to:', state.currentTrack);

    load(`/api/track/${state.currentTrack!.id}?Access-Token=${authState.token ?? ''}`);
    play();
    setIsPlaying(true);

    audio.addEventListener('ended', () => {
      console.log('track ended');
      dispatch({ type: 'NEXT_TRACK', payload: null });
    }, { once: true });

    return () => {
      pause();
    }
  }, [dispatch, state.currentTrack, authState.token, playedGarbage, setPlayedGarbage]);

  // Monitor pause/play changes.
  useEffect(() => {
    if (isPlaying) {
      console.log('play...');
      play();
    } else {
      pause();
    }
  }, [isPlaying])

  return (
    <div className={classnames(styles.player /*, { [styles.hidden]: !state.currentTrack }*/)}>
      <h1>{state.currentTrack ? state.currentTrack.name : '...'}</h1>
      <div className={styles.controls}>
        <PlayerControls
          isPlaying={isPlaying}
          onPausePressed={() => setIsPlaying(false)}
          onPlayPressed={() => setIsPlaying(true)}
          onNextPressed={() => {
            setIsPlaying(false);
            dispatch({ type: 'NEXT_TRACK', payload: null });
          }}
          onPreviousPressed={() => {
            setIsPlaying(false);
            dispatch({ type: 'PREVIOUS_TRACK', payload: null });
          }}
        />
        <PlayerTime audioPlayerRef={audio} />
      </div>
    </div>
  );
};

export default Player;