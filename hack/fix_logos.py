import sys
import os

filepath = "internal/pkg/defaulting/applicationcatalog.go"
temp_filepath = filepath + ".tmp"

with open(filepath, 'r') as f_in, open(temp_filepath, 'w') as f_out:
    in_logo = False
    logo_buffer = ""
    prefix = ""
    
    for line in f_in:
        stripped_line = line.strip()
        
        if in_logo:
            # We are inside a broken string
            if stripped_line.endswith('",'):
                # Found the end
                content = stripped_line[:-2] # remove ",
                logo_buffer += content
                # Write the full line
                f_out.write(prefix + logo_buffer + '",\n')
                in_logo = False
                logo_buffer = ""
                prefix = ""
            else:
                # Middle part
                logo_buffer += stripped_line
        else:
            # Check if this line starts a broken logo string
            # It must contain 'Logo:' and '"' but NOT end with '",'
            # Note: A valid one-line logo ends with '",'
            
            if 'Logo:' in line and '"' in line:
                # Check if it ends with ", (ignoring whitespace)
                if not stripped_line.endswith('",'):
                    # It is broken
                    in_logo = True
                    # Find start of string
                    # Line looks like: \t\t\t\tLogo:             "BASE64...
                    # We want to keep indentation and "Logo:             " part as prefix
                    quote_index = line.find('"')
                    prefix = line[:quote_index+1]
                    # buffer is everything after the first quote
                    logo_buffer = line[quote_index+1:].strip() 
                    # If the line ended with a newline char that caused the break, strip() handles it.
                    # But wait, line.strip() removes trailing newline.
                    # The content on this line is line[quote_index+1:].strip() (removes \n at end)
                    pass
                else:
                    # It's a valid one-liner (or empty string ""), just write it
                    f_out.write(line)
            else:
                f_out.write(line)

os.replace(temp_filepath, filepath)
print("Fixed newlines in Logo strings.")
