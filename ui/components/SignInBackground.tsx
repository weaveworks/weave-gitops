// @ts-nocheck
import React from "react";
import Lottie from "react-lottie-player";
import SignInBackground from "../images/SignInBackground.json";

const SignInbackground = () => {
  return (
    <Lottie
      loop
      animationData={SignInBackground}
      play
      style={{
        width: "100%",
        position: "absolute",
        zIndex: -999,
      }}
    />
  );
};

export default SignInbackground;
