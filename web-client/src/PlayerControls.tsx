import React from 'react';
import nextIcon from './svg/next.svg';
import pauseIcon from './svg/pause.svg';
import playIcon from './svg/play.svg';
import previousIcon from './svg/previous.svg';

export interface PlayerControlsProps {
  isPlaying: boolean;
  onPausePressed: () => void;
  onPlayPressed: () => void;
  onNextPressed: () => void;
  onPreviousPressed: () => void;
}

const PlayerControls = ({ isPlaying, onPausePressed, onPlayPressed, onNextPressed, onPreviousPressed }: PlayerControlsProps) => {
  const PauseButton = () => {
    return (
      <span role="button" onClick={onPausePressed}><img src={pauseIcon} alt="pause button" /></span>
    );
  };

  const PlayButton = () => {
    return (
      <span role="button" onClick={onPlayPressed}><img src={playIcon} alt="play button" /></span>
    );
  };

  const NextButton = () => {
    return (
      <span role="button" onClick={onNextPressed}><img src={nextIcon} alt="next button" /></span>
    );
  }

  const PreviousButton = () => {
    return (
      <span role="button" onClick={onPreviousPressed}><img src={previousIcon} alt="previous button" /></span>
    );
  }

  return (
    <>
      <PreviousButton />
      {isPlaying ? <PauseButton /> : <PlayButton />}
      <NextButton />
    </>
  )
};

export default PlayerControls;
