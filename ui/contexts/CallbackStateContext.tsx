import * as React from "react";
import { CallbackSessionState } from "../lib/storage";

type Props = {
  callbackState?: CallbackSessionState;
  children?: any;
};

export type CallbackStateContextType = {
  callbackState: CallbackSessionState;
};
export const CallbackStateContext =
  React.createContext<CallbackStateContextType>(null);

export default function CallbackStateContextProvider({
  callbackState,
  children,
}: Props) {
  const value: CallbackStateContextType = {
    callbackState,
  };
  return <CallbackStateContext.Provider value={value} children={children} />;
}
