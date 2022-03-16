import { Meta, Story } from "@storybook/react";
import React from "react";
import SearchInput, { Props } from "../components/SearchInput";

export default {
  title: "SearchInput",
  component: SearchInput,
  parameters: {
    docs: {
      description: {
        component:
          "Series of deletable MUI Chip components: https://mui.com/components/chips/",
      },
    },
  },
} as Meta;

const Template: Story<Props> = (args) => {
  return <SearchInput {...args} />;
};

export const Default = Template.bind({});
Default.args = {};
