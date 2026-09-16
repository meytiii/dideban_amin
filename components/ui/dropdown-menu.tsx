"use client";

import { Button } from "./button";
import { ChevronDown } from "lucide-react";
import React, { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";

export type DropdownOption = {
  label: string;
  onClick: () => void;
  Icon?: React.ReactNode;
};

type DropdownMenuProps = {
  options: DropdownOption[];
  children?: React.ReactNode;
  className?: string;
};

const DropdownMenu = ({ options, children, className = "" }: DropdownMenuProps) => {
  const [isOpen, setIsOpen] = useState(false);

  const toggleDropdown = () => {
    setIsOpen(!isOpen);
  };

  return (
    <div className={`relative ${className}`}>
      <Button
        onClick={toggleDropdown}
        className="px-4 py-2 bg-[#11111198] hover:bg-[#111111d1] shadow-[0_0_20px_rgba(0,0,0,0.2)] border-none rounded-xl backdrop-blur-sm flex items-center justify-between gap-2"
      >
        <span>{children ?? "Menu"}</span>
        <motion.span
          animate={{ rotate: isOpen ? 180 : 0 }}
          transition={{ duration: 0.4, ease: "easeInOut", type: "spring" }}
        >
          <ChevronDown className="h-4 w-4" />
        </motion.span>
      </Button>

      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ y: -5, scale: 0.95, filter: "blur(10px)", opacity: 0 }}
            animate={{ y: 0, scale: 1, filter: "blur(0px)", opacity: 1 }}
            exit={{ y: -5, scale: 0.95, opacity: 0, filter: "blur(10px)" }}
            transition={{ duration: 0.4, ease: "circInOut" }}
            className="absolute z-50 w-64 mt-2 p-1.5 bg-[#0f172ae6] border border-blue-500/30 rounded-xl shadow-[0_10px_30px_rgba(0,0,0,0.5)] backdrop-blur-md flex flex-col gap-1"
          >
            {options && options.length > 0 ? (
              options.map((option, index) => (
                <motion.button
                  initial={{
                    opacity: 0,
                    x: 10,
                    scale: 0.95,
                    filter: "blur(6px)",
                  }}
                  animate={{ opacity: 1, x: 0, scale: 1, filter: "blur(0px)" }}
                  exit={{
                    opacity: 0,
                    x: 10,
                    scale: 0.95,
                    filter: "blur(6px)",
                  }}
                  transition={{
                    duration: 0.3,
                    delay: index * 0.04,
                    ease: "easeInOut",
                  }}
                  whileHover={{
                    backgroundColor: "rgba(59, 130, 246, 0.2)",
                    x: -2,
                    transition: {
                      duration: 0.2,
                      ease: "easeInOut",
                    },
                  }}
                  whileTap={{
                    scale: 0.97,
                    transition: {
                      duration: 0.1,
                      ease: "easeInOut",
                    },
                  }}
                  key={option.label}
                  onClick={() => {
                    option.onClick();
                    setIsOpen(false);
                  }}
                  className="px-3 py-2.5 cursor-pointer text-white text-sm rounded-lg w-full text-right flex items-center gap-x-2.5 transition-colors"
                >
                  {option.Icon}
                  <span>{option.label}</span>
                </motion.button>
              ))
            ) : (
              <div className="px-4 py-2 text-slate-400 text-xs">موردی یافت نشد</div>
            )}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};

export { DropdownMenu };
