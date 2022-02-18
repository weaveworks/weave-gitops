import { Meta, Story } from "@storybook/react";
import * as React from "react";
import DataTable, { Props, SortType } from "../components/DataTable";

const rows = [
  {
    name: "the-cool-app",
    status: "Ready",
    lastUpdate: "2005-01-02T15:04:05-0700",
    lastSyncedAt: 1000,
  },
  {
    name: "podinfo",
    status: "Failed",
    lastUpdate: "2006-01-02T15:04:05-0700",
    lastSyncedAt: 2000,
  },
  {
    name: "nginx",
    status: "Ready",
    lastUpdate: "2004-01-02T15:04:05-0700",
    lastSyncedAt: 3000,
  },
];

const fields = [
  {
    label: "Name",
    value: ({ name }) => <a href="/some_url">{name}</a>,
    sortType: SortType.string,
    altSortValue: ({ name }) => name,
  },
  {
    label: "Status",
    value: "status",
    sortType: SortType.bool,
    altSortValue: ({ status }) => (status === "Ready" ? true : false),
  },
  {
    label: "Last Updated",
    value: "lastUpdate",
    sortType: SortType.date,
    altSortValue: ({ lastUpdate }) => lastUpdate,
  },
  {
    label: "Last Synced At",
    value: "lastSyncedAt",
    sortType: SortType.number,
  },
];

export default {
  title: "DataTable",
  component: DataTable,
  parameters: {
    docs: {
      description: {
        component: "Generates a sorted table based on given fields and rows",
      },
    },
  },
} as Meta;

const Template: Story<Props> = (args) => <DataTable {...args} />;

export const WithData = Template.bind({});
WithData.args = {
  fields: fields,
  rows: rows,
  defaultSort: fields[0],
  className: "",
};
export const NoRows = Template.bind({});
NoRows.args = { ...WithData.args, rows: [] };
