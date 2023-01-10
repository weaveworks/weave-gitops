import React from "react";
import DataTable from "../DataTable";

const ImageAutomationsTable = () => {
  const data = [
    {
      Name: "my app",
      Namespace: "flux-system",
      Source: "main",
      LastCommit: "abc123",
      lastUpdatedAt: Date.now(),
    },
    {
      Name: "my app 2",
      Namespace: "flux-system",
      Source: "dev",
      LastCommit: "abc1231212",
      lastUpdatedAt: Date.now(),
    }
  ];
  const initialFilterState = {
    // ...filterConfig(versions, "version"),
  };
  return (
    <DataTable
      // className={className}
      filters={initialFilterState}
      rows={data}
      fields={[
        {
          label: "Cluster",
          value: "clusterName",
          textSearchable: true,
          maxWidth: 600,
        },
        {
          label: "Namespace",
          value: "namespace",
        },
        {
          label: "Flux Version",
          value: "version",
        },
      ]}
    />
  );
};

export default ImageAutomationsTable;
