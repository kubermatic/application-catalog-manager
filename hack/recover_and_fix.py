import re
import os

filepath = "internal/pkg/defaulting/applicationcatalog.go"
temp_filepath = filepath + ".recovered"

# Regex to match the messed up line
# \s+Logo:\s*".*LogoFormat:\s*".*",
pattern = re.compile(r'^(\s+)Logo:\s*"(.*)LogoFormat:\s*"([^"]+)",\s*$')

with open(filepath, 'r') as f_in, open(temp_filepath, 'w') as f_out:
    for line in f_in:
        match = pattern.match(line)
        if match:
            indent = match.group(1)
            logo_content = match.group(2)
            format_content = match.group(3)
            
            # The logo_content is the base64 string.
            # It shouldn't contain newlines if my previous script did its job of joining.
            # But just in case, let's clean it up (though regex .* is greedy, it stops at LogoFormat).
            
            # Write Logo line
            f_out.write(f'{indent}Logo:             "{logo_content}",\n')
            # Write LogoFormat line (aligning as per convention)
            f_out.write(f'{indent}LogoFormat:       "{format_content}",\n')
        else:
            f_out.write(line)

os.replace(temp_filepath, filepath)
print("Recovered Logo and LogoFormat lines.")
