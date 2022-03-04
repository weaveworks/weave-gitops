import { Meta, Story } from "@storybook/react";
import React from "react";
import FilterDialog, { FilterConfig, Props } from "../components/FilterDialog";
import Flex from "../components/Flex";

export default {
  title: "FilterBar",
  component: FilterDialog,
  parameters: {
    docs: {
      description: {
        component: "Filters row data to be passed down to `<DataTable />`",
      },
    },
  },
} as Meta;

const Template: Story<Props> = (args) => {
  const [storyFilters, setStoryFilters] = React.useState<FilterConfig>(
    args.filterList
  );

  return (
    <Flex wide align end>
      <FilterDialog
        {...args}
        filterList={storyFilters}
        onFilterSelect={(v) => setStoryFilters(v)}
      />
    </Flex>
  );
};

export const Default = Template.bind({});
Default.args = {
  filterList: {
    Name: ["app", "app2", "app3"],
    Status: ["Ready", "Failed"],
    Type: ["Application", "Helm Release"],
  },
};
