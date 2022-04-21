import React, { useEffect, useState } from "react";
import LottiePlayerLight from "react-lottie-player/dist/LottiePlayerLight";

const LottieWrapper = () => {
  const [animationData, setAnimationData] = useState<any>();

  useEffect(() => {
    import(`../../images/SignInBackground.json`).then(setAnimationData);
  }, []);

  return (
    <LottiePlayerLight
      play
      loop={false}
      speed={0.3}
      animationData={animationData}
      // renderer="canvas"
      rendererSettings={{ preserveAspectRatio: "xMidYMid slice" }}
      style={{
        width: "100%",
        height: "100%",
        position: "absolute",
        zIndex: -999,
      }}
    />
  );
};
export default LottieWrapper;
