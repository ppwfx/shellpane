# shellpane

- BE
    - default values
    - add view description
      - view add steps, e.g. to verify and confirm
    - output stderr
    - add tests
    - add width
    - permissions
  
    - steps
      - error highlight
    
- FE
    - description
    - autocomplete
    - download extension
    - always print stdout and stderr!
    - print statuscode


```
env:
- name: DIR
  validator: `^[A-Za-z0-9_./-]{1,15}$`
  validate:
    mustMatch: `^[A-Za-z0-9_./-]{1,15}$`
    isRequired: true
    isNumeric: true