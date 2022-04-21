import React, { useEffect, useState } from "react";
import Lottie from "react-lottie-player";

const AnimatedBackground = () => {
  const [animationData, setAnimationData] = useState<any>();

  useEffect(() => {
    import(`../images/SignInBackground.json`).then(setAnimationData);
  }, []);

  // if (!animationData) return <div>Loading...</div>;

  return (
    <Lottie
      play
      loop={false}
      speed={0.3}
      animationData={animationData}
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
export default AnimatedBackground;
