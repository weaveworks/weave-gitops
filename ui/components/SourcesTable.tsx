import * as React from "react";
import styled from "styled-components";
import { Button } from "..";
import { Source, V2Routes } from "../lib/types";
import { formatAppScopedURL, formatURL } from "../lib/utils";
import DataTable from "./DataTable";
import Flex from "./Flex";
import Link from "./Link";

type Props = {
  className?: string;
  sources: Source[];
  appName?: string;
};

function SourcesTable({ className, sources, appName }: Props) {
  return (
    <div className={className}>
      <Flex align wide between>
        <h3>Sources</h3>
        <Link
          to={
            appName
              ? formatAppScopedURL(appName, V2Routes.AddSource)
              : formatURL(V2Routes.AddSource)
          }
        >
          <Button>Add Source</Button>
        </Link>
      </Flex>

      <DataTable
        sortFields={["name"]}
        rows={sources}
        fields={[
          {
            label: "Name",
            value: (k) => (
              <Link
                to={formatURL(V2Routes.Source, {
                  name: k.name,
                  namespace: k.namespace,
                })}
              ></Link>
            ),
          },
          { label: "Type", value: "type" },
          { label: "Namespace", value: "namespace" },
        ]}
      />
    </div>
  );
}

export default styled(SourcesTable).attrs({ className: SourcesTable.name })``;
