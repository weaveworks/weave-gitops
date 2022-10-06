import * as React from "react";
import { formatURL } from "../lib/nav";

export type LinkResolver = (str: string, params?: any) => string;

interface Props {
  resolver: LinkResolver;
  children: any;
}

function defaultResolver(str: string, params?: any) {
  return formatURL(str, params);
}

const LinkResolverContext = React.createContext<LinkResolver>(defaultResolver);

export function LinkResolverProvider({ resolver, children }: Props) {
  return (
    <LinkResolverContext.Provider value={resolver}>
      {children}
    </LinkResolverContext.Provider>
  );
}

export function useLinkResolver() {
  return React.useContext(LinkResolverContext);
}
