# shellpane

- BE
    - specs.yaml -> shellpane.yaml
    - specs.yaml makes specs a field
    - add input validation
      
    - default values
    - add view description
      - view add steps, e.g. to verify and confirm
    - add tags
      - should be possible to define a color
    - output stderr
    - add tests
    - add width
    - permissions
    - steps
    
- FE
    - hide command
    - description
    - scroll to top after command execution
    - autocomplete
    - download extension
    - disable input when loading
    - update indicator
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