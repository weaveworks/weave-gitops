import React, { useEffect, useState } from "react";
import Lottie from "react-lottie-player/dist/LottiePlayerLight";

const LottieWrapper = () => {
  const [animationData, setAnimationData] = useState<any>();

  useEffect(() => {
    import(`../../images/error404.json`).then(setAnimationData);
  }, []);

  return (
    <Lottie loop animationData={animationData} play style={{ height: 650 }} />
  );
};
export default LottieWrapper;
