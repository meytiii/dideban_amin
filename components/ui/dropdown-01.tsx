import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { ChevronDown, Moon, Sun, Check } from 'lucide-react';

const DropdownComponent = () => {
  const [isDark, setIsDark] = useState(true);
  const [isOpen, setIsOpen] = useState(false);
  const [selected, setSelected] = useState('Select an option');

  const options = [
    { id: 1, label: 'Premium Plan', description: 'Advanced features' },
    { id: 2, label: 'Professional Plan', description: 'For power users' },
    { id: 3, label: 'Enterprise Plan', description: 'Unlimited access' },
    { id: 4, label: 'Starter Plan', description: 'Get started free' },
  ];

  const handleSelect = (option: typeof options[0]) => {
    setSelected(option.label);
    setIsOpen(false);
  };

  return (
    <div className={`min-h-screen w-full transition-colors duration-500 ${
      isDark ? 'bg-black' : 'bg-white'
    } relative overflow-hidden`}>
      <div className="absolute inset-0 opacity-20">
        <svg width="100%" height="100%" xmlns="http://www.w3.org/2000/svg">
          <defs>
            <pattern id="grid" width="40" height="40" patternUnits="userSpaceOnUse">
              <path
                d="M 40 0 L 0 0 0 40"
                fill="none"
                stroke={isDark ? '#ffffff' : '#000000'}
                strokeWidth="0.5"
              />
            </pattern>
          </defs>
          <rect width="100%" height="100%" fill="url(#grid)" />
        </svg>
      </div>

      <motion.button
        whileHover={{ scale: 1.05 }}
        whileTap={{ scale: 0.95 }}
        onClick={() => setIsDark(!isDark)}
        className={`fixed top-8 right-8 p-3 rounded-full border ${
          isDark 
            ? 'bg-white text-black border-white' 
            : 'bg-black text-white border-black'
        } shadow-lg z-50`}
      >
        {isDark ? <Sun className="w-5 h-5" /> : <Moon className="w-5 h-5" />}
      </motion.button>

      <div className="relative z-10 flex items-center justify-center min-h-screen px-4">
        <div className="w-full max-w-md space-y-6">
          <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6 }}
            className="text-center space-y-2"
          >
            <h1 className={`text-4xl font-bold tracking-tight ${
              isDark ? 'text-white' : 'text-black'
            }`}>
              Select Dropdown
            </h1>
            <p className={`text-sm ${
              isDark ? 'text-gray-400' : 'text-gray-600'
            }`}>
              Select your preferred option
            </p>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.5, delay: 0.2 }}
            className="relative"
          >
            <motion.button
              whileHover={{ scale: 1.01 }}
              whileTap={{ scale: 0.99 }}
              onClick={() => setIsOpen(!isOpen)}
              className={`w-full px-6 py-4 rounded-xl border-2 transition-all duration-300 ${
                isDark
                  ? 'bg-black border-white text-white hover:bg-transparent hover:text-white'
                  : 'bg-white border-black text-black hover:bg-black hover:text-white'
              } flex items-center justify-between group`}
            >
              <span className="font-medium">{selected}</span>
              <motion.div
                animate={{ rotate: isOpen ? 180 : 0 }}
                transition={{ duration: 0.3 }}
              >
                <ChevronDown className="w-5 h-5" />
              </motion.div>
            </motion.button>

            <AnimatePresence>
              {isOpen && (
                <motion.div
                  initial={{ opacity: 0, y: -10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -10 }}
                  transition={{ duration: 0.2 }}
                  className={`absolute w-full mt-2 rounded-xl border-2 overflow-hidden shadow-2xl ${
                    isDark
                      ? 'bg-black border-white'
                      : 'bg-white border-black'
                  }`}
                >
                  {options.map((option, index) => (
                    <motion.button
                      key={option.id}
                      initial={{ opacity: 0, x: -20 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ duration: 0.2, delay: index * 0.05 }}
                      onClick={() => handleSelect(option)}
                      className={`w-full px-6 py-4 text-left transition-colors duration-200 ${
                        isDark
                          ? 'hover:bg-white hover:text-black text-white'
                          : 'hover:bg-black hover:text-white text-black'
                      } flex items-center justify-between group ${
                        index !== options.length - 1
                          ? isDark
                            ? 'border-b border-gray-800'
                            : 'border-b border-gray-200'
                          : ''
                      }`}
                    >
                      <div>
                        <div className="font-medium">{option.label}</div>
                        <div className={`text-sm ${
                          isDark ? 'text-gray-400 group-hover:text-gray-600' : 'text-gray-600 group-hover:text-gray-400'
                        }`}>
                          {option.description}
                        </div>
                      </div>
                      {selected === option.label && (
                        <motion.div
                          initial={{ scale: 0 }}
                          animate={{ scale: 1 }}
                          transition={{ type: "spring", stiffness: 500, damping: 30 }}
                        >
                          <Check className="w-5 h-5" />
                        </motion.div>
                      )}
                    </motion.button>
                  ))}
                </motion.div>
              )}
            </AnimatePresence>
          </motion.div>

          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.5, delay: 0.4 }}
            className={`text-center text-sm ${
              isDark ? 'text-gray-500' : 'text-gray-500'
            }`}
          >
            Current selection: <span className="font-medium">{selected}</span>
          </motion.div>
        </div>
      </div>
    </div>
  );
};

export default DropdownComponent;
