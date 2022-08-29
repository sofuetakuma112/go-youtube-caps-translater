from nemo.collections import nlp as nemo_nlp

import sys
import os

if __name__ == "__main__":
    args = sys.argv
    videoId = args[1]
    captionTxtFileName = args[2]
    restoredPuncTxtFileName = args[3]

    currentDir = os.getcwd()
    targetDir = f"{currentDir}/captions/{videoId}"
    f = open(f"{targetDir}/{captionTxtFileName}", "r")

    text = f.read()
    f.close()

    pretrained_model = nemo_nlp.models.PunctuationCapitalizationModel.from_pretrained("punctuation_en_bert")

    inference_results = pretrained_model.add_punctuation_capitalization(
        [
            text
        ],
        max_seq_length=128,
        step=8,
        margin=16,
        batch_size=32,
    )

    f = open(f"{targetDir}/{restoredPuncTxtFileName}", "w")
    f.write(inference_results[0])
    f.close()
