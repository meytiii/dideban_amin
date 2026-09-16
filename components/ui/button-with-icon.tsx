import React from "react";
import { Button } from "./button";
import { ArrowUpRight } from "lucide-react";

interface ButtonWithIconProps {
  label?: string;
  icon?: React.ReactNode;
  onClick?: () => void;
  className?: string;
}

const ButtonWithIconDemo: React.FC<ButtonWithIconProps> = ({
  label = "Let's Collaborate",
  icon = <ArrowUpRight size={16} />,
  onClick,
  className = "",
}) => {
  return (
    <Button
      onClick={onClick}
      className={`relative text-sm font-medium rounded-full h-12 p-1 ps-6 pe-14 group transition-all duration-500 hover:ps-14 hover:pe-6 w-fit overflow-hidden cursor-pointer ${className}`}
    >
      <span className="relative z-10 transition-all duration-500">
        {label}
      </span>
      <div className="absolute right-1 w-10 h-10 bg-background text-foreground rounded-full flex items-center justify-center transition-all duration-500 group-hover:right-[calc(100%-44px)] group-hover:rotate-45">
        {icon}
      </div>
    </Button>
  );
};

export default ButtonWithIconDemo;
