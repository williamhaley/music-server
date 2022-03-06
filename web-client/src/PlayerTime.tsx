import React, { useEffect, useMemo, useState } from 'react';
import { DateTime } from 'luxon';

export interface PlayerTimeProps {
  audioPlayerRef: HTMLAudioElement | null;
}

const PlayerTime = ({ audioPlayerRef }: PlayerTimeProps) => {
  const [currentTime, setCurrentTime] = useState(0);
  const [totalTime, setTotalTime] = useState(0);

  const onDurationChange = (event: Event) => {
    if (!event.target) {
      return;
    }

    const asAudioEl = event.target as HTMLAudioElement;
    setTotalTime(asAudioEl.duration);
  };

  // Do this here so we don't deal with re-renders in the parent.
  // This is constantly triggering state updates.
  useEffect(() => {
    if (audioPlayerRef === null) {
      return;
    }

    // Constantly check the current time.
    const timer = setInterval(() => {
      if (!audioPlayerRef.currentTime) {
        return;
      }

      setCurrentTime(audioPlayerRef.currentTime);
    }, 10);

    audioPlayerRef.addEventListener('durationchange', onDurationChange);

    return () => {
      clearInterval(timer);
      audioPlayerRef.removeEventListener('durationchange', onDurationChange)
    };
  }, [setTotalTime, audioPlayerRef]);

  const formattedTotalTime = useMemo(() => {
    if (!totalTime) {
      return '--';
    }

    return DateTime.fromSeconds(totalTime).toFormat('mm:ss')
  }, [totalTime]);

  const formattedCurrentTime = useMemo(() => {
    if (!currentTime) {
      return '--';
    }

    return DateTime.fromSeconds(currentTime).toFormat('mm:ss')
  }, [currentTime]);

  return (
    <div>{formattedCurrentTime} / {formattedTotalTime}</div>
  );
};

export default PlayerTime;