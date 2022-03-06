const audio = new Audio();

// Safari is a real pain and we need an unbroken (more or less) promise chain from a user clicking something
// to us playing audio. This empty track allows us to say, "Ah, yes, right away! We will play something!", but it
// is dead air. This gives us an in though, and we can programatically play tracks going forward.
// TODO WFH Check if we already did a dummy play at least once. If so, skip subsequent calls.
const dummyPlay = () => {
  audio.src = "data:audio/mpeg;base64,SUQzBAAAAAABEVRYWFgAAAAtAAADY29tbWVudABCaWdTb3VuZEJhbmsuY29tIC8gTGFTb25vdGhlcXVlLm9yZwBURU5DAAAAHQAAA1N3aXRjaCBQbHVzIMKpIE5DSCBTb2Z0d2FyZQBUSVQyAAAABgAAAzIyMzUAVFNTRQAAAA8AAANMYXZmNTcuODMuMTAwAAAAAAAAAAAAAAD/80DEAAAAA0gAAAAATEFNRTMuMTAwVVVVVVVVVVVVVUxBTUUzLjEwMFVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVf/zQsRbAAADSAAAAABVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVf/zQMSkAAADSAAAAABVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVV";

  return new Promise((resolve) => {
    audio.addEventListener('ended', () => {
      resolve(true);
    }, { once: true });

    audio.play();
  });
}

const play = () => {
  audio.play();
};

const load = (src: string) => {
  // Pause before changing the src to prevent the audio play interpreting it as an abort of the current request.
  audio.pause();
  audio.src = src;
};

const pause = () => {
  audio.pause();
};

export { play, load, dummyPlay, pause, audio };