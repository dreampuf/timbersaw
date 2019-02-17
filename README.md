# HTTP log monitoring console program

  
![Program Structure](timbersaw.png)

## Main Features

- Living monitor for local files.
- Extendable format support.
- Test cases coverage major part.

This little console application demonstrates how to gracefully handle log files and provide extendability for future improvement.

## Defect and Improvement

1. No continuous processing support after the app restart. It reprocesses complete data every time.
   
   We can add a position memory function to mark where it has been processed.
   
2. A long period of use case might cause a memory issue.

   I'm using a `container/ring` as the data structure to compute how much status remain. Considering two sides, it's a little bit costly. 
   On the memory cost side, it requires a ring bucket object per second(in our case). 
   On the runtime cost side, it's inefficient. Many operators just for typecasting, it can be improved by developing a type related ring.
