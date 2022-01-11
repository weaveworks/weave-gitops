import { useQuery } from "react-query";
import { Source } from "../lib/api/app/source.pb";

export function useListSources() {
  const sources: Source[] = [
    {
      name: "podinfo",
      url: "git@github.com:stefanprodan/podinfo.git",
      reference: {
        branch: "main",
      },
      provider: "github",
    },
  ];

  return useQuery(
    "sources",
    () => new Promise<Source[]>((accept) => accept(sources))
  );
}
