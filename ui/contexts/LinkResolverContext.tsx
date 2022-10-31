import * as React from "react";

export type LinkResolver = (str: string, params?: any) => string;

interface Props {
  resolver: LinkResolver;
  children: any;
}

const LinkResolverContext = React.createContext<LinkResolver>(null);

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
