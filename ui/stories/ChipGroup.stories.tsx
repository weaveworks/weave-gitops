import { Meta, Story } from "@storybook/react";
import React from "react";
import ChipGroup, { Props } from "../components/ChipGroup";

export default {
  title: "ChipGroup",
  component: ChipGroup,
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
  const [storyChips, setStoryChips] = React.useState([
    "chip",
    "chippy",
    "another one",
  ]);

  return (
    <ChipGroup {...args} chips={storyChips} onChipRemove={setStoryChips} />
  );
};

export const Default = Template.bind({});
Default.args = {};
