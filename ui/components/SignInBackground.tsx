// @ts-nocheck
import React from "react";
import Lottie from "react-lottie-player";
import SignInBackground from "../images/SignInBackground.json";

const SignInbackground = () => {
  return (
    <Lottie
      autoPlay
      loop
      animationData={SignInBackground}
      style={{
        width: "100%",
        position: "absolute",
        zIndex: -999,
      }}
    />
  );
};

export default SignInbackground;
