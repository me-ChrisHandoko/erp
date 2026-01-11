"use client";

import * as React from "react";
import { Check, ChevronsUpDown } from "lucide-react";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";

export interface ComboboxOption {
  value: string;
  label: string;
  searchLabel?: string;
  disabled?: boolean;
  group?: string;
}

interface ComboboxProps {
  options: ComboboxOption[];
  value?: string;
  onValueChange?: (value: string) => void;
  placeholder?: string;
  searchPlaceholder?: string;
  emptyMessage?: string;
  disabled?: boolean;
  className?: string;
  renderOption?: (option: ComboboxOption, selected: boolean) => React.ReactNode;
  renderTrigger?: (
    selectedOption: ComboboxOption | undefined
  ) => React.ReactNode;
}

export function Combobox({
  options,
  value,
  onValueChange,
  placeholder = "Select option...",
  searchPlaceholder = "Search...",
  emptyMessage = "No results found.",
  disabled = false,
  className,
  renderOption,
  renderTrigger,
}: ComboboxProps) {
  const [open, setOpen] = React.useState(false);
  const commandListRef = React.useRef<HTMLDivElement>(null);
  const inputRef = React.useRef<HTMLInputElement>(null);

  // Autofocus input when popover opens
  React.useEffect(() => {
    if (open) {
      // Longer delay to ensure popover animation completes and DOM is ready
      const timer = setTimeout(() => {
        const input = inputRef.current;
        if (input) {
          input.focus();
        } else {
          // Fallback: try to find input by querySelector if ref doesn't work
          const commandInput = document.querySelector('[cmdk-input]') as HTMLInputElement;
          if (commandInput) {
            commandInput.focus();
          }
        }
      }, 100);

      return () => clearTimeout(timer);
    }
  }, [open]);

  // Fix scroll in Dialog: Add wheel event listener with multiple approaches
  React.useEffect(() => {
    if (!open) return;

    const listElement = commandListRef.current;
    if (!listElement) return;

    // Also try to get the actual scrollable element via querySelector
    const scrollableElement = listElement.querySelector('[cmdk-list]') as HTMLElement || listElement;

    const handleWheel = (e: WheelEvent) => {
      const target = scrollableElement || listElement;
      const { scrollTop, scrollHeight, clientHeight } = target;
      const isAtTop = scrollTop === 0;
      const isAtBottom = scrollTop + clientHeight >= scrollHeight - 1;

      const scrollingUp = e.deltaY < 0;
      const scrollingDown = e.deltaY > 0;

      // Prevent default only if we can scroll within the list
      if ((scrollingUp && !isAtTop) || (scrollingDown && !isAtBottom)) {
        e.preventDefault();
      }

      e.stopPropagation();
      target.scrollTop += e.deltaY;
    };

    // Attach to both elements for safety
    scrollableElement.addEventListener("wheel", handleWheel, { passive: false });
    if (scrollableElement !== listElement) {
      listElement.addEventListener("wheel", handleWheel, { passive: false });
    }

    return () => {
      scrollableElement.removeEventListener("wheel", handleWheel);
      if (scrollableElement !== listElement) {
        listElement.removeEventListener("wheel", handleWheel);
      }
    };
  }, [open]);

  const selectedOption = options.find((option) => option.value === value);

  // Group options by group property
  const groupedOptions = React.useMemo(() => {
    const grouped: Record<string, ComboboxOption[]> = {};
    const ungrouped: ComboboxOption[] = [];

    options.forEach((option) => {
      if (option.group) {
        if (!grouped[option.group]) {
          grouped[option.group] = [];
        }
        grouped[option.group].push(option);
      } else {
        ungrouped.push(option);
      }
    });

    return { grouped, ungrouped };
  }, [options]);

  const handleSelect = (selectedValue: string) => {
    const newValue = selectedValue === value ? "" : selectedValue;
    onValueChange?.(newValue);
    setOpen(false);
  };

  return (
    <Popover open={open} onOpenChange={setOpen} modal={true}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          disabled={disabled}
          className={cn(
            "w-full justify-between",
            !value && "text-muted-foreground",
            className
          )}
        >
          {renderTrigger ? (
            renderTrigger(selectedOption)
          ) : (
            <span className="truncate">
              {selectedOption ? selectedOption.label : placeholder}
            </span>
          )}
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent
        className="p-0 popover-content-width-same-as-trigger"
        align="start"
        style={{
          width: "var(--radix-popover-trigger-width)",
          overscrollBehavior: "contain"
        }}
        onOpenAutoFocus={(e) => {
          e.preventDefault();
          // Focus the input after preventing default
          setTimeout(() => {
            const input = inputRef.current || document.querySelector('[cmdk-input]') as HTMLInputElement;
            input?.focus();
          }, 0);
        }}
      >
        <Command
          className="w-full"
          loop
          filter={(value, search) => {
            // Custom filter: strict substring matching instead of fuzzy search
            if (!search) return 1;
            const normalizedValue = value.toLowerCase();
            const normalizedSearch = search.toLowerCase();

            // Prioritize starts-with matches
            if (normalizedValue.startsWith(normalizedSearch)) return 1;

            // Allow contains matches with lower priority
            if (normalizedValue.includes(normalizedSearch)) return 0.8;

            // No match
            return 0;
          }}
        >
          <CommandInput ref={inputRef} placeholder={searchPlaceholder} />
          <CommandList
            ref={commandListRef}
            className="max-h-[300px] overflow-y-auto overscroll-contain"
            style={{
              scrollPaddingBlock: "8px",
              overscrollBehavior: "contain",
              WebkitOverflowScrolling: "touch"
            }}
          >
            <CommandEmpty>{emptyMessage}</CommandEmpty>

            {/* Render ungrouped options */}
            {groupedOptions.ungrouped.length > 0 && (
              <CommandGroup>
                {groupedOptions.ungrouped.map((option) => {
                  const isSelected = option.value === value;
                  return (
                    <CommandItem
                      key={option.value}
                      value={option.searchLabel || option.label}
                      disabled={option.disabled}
                      onSelect={() =>
                        !option.disabled && handleSelect(option.value)
                      }
                    >
                      {renderOption ? (
                        renderOption(option, isSelected)
                      ) : (
                        <>
                          <Check
                            className={cn(
                              "mr-2 h-4 w-4",
                              isSelected ? "opacity-100" : "opacity-0"
                            )}
                          />
                          {option.label}
                        </>
                      )}
                    </CommandItem>
                  );
                })}
              </CommandGroup>
            )}

            {/* Render grouped options */}
            {Object.entries(groupedOptions.grouped).map(
              ([groupName, groupOptions]) => (
                <CommandGroup key={groupName} heading={groupName}>
                  {groupOptions.map((option) => {
                    const isSelected = option.value === value;
                    return (
                      <CommandItem
                        key={option.value}
                        value={option.searchLabel || option.label}
                        disabled={option.disabled}
                        onSelect={() =>
                          !option.disabled && handleSelect(option.value)
                        }
                      >
                        {renderOption ? (
                          renderOption(option, isSelected)
                        ) : (
                          <>
                            <Check
                              className={cn(
                                "mr-2 h-4 w-4",
                                isSelected ? "opacity-100" : "opacity-0"
                              )}
                            />
                            {option.label}
                          </>
                        )}
                      </CommandItem>
                    );
                  })}
                </CommandGroup>
              )
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
